package tgfun

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	swissknife "github.com/Sagleft/swiss-knife"
	tb "gopkg.in/telebot.v3"
)

func getFilePath(localPath, pathRoot string) string {
	filePath := pathRoot
	if !strings.HasSuffix(filePath, "/") {
		filePath += "/"
	}

	filePath += localPath
	return filePath
}

func getMessageType(message EventMessage) MessageType {
	if message.Image != "" {
		return MessageTypePhoto
	}

	if message.File.Path != "" {
		return MessageTypeDocument
	}

	if message.Video.Path != "" {
		return MessageTypeVideo
	}

	if message.Audio.Path != "" {
		return MessageTypeAudio
	}

	return MessageTypeText
}

// returns photo, state
func (q *QueryHandler) getPhotoMessage(
	message EventMessage,
	filesRoot string,
	telegramUserID int64,
) (interface{}, fileState) {
	st := fileState{
		Type:          MessageTypePhoto,
		LocalFilePath: message.Image,
	}

	photo := &tb.Photo{}
	if strings.Contains(message.Image, "http") {
		// external image
		photo.File = tb.FromURL(message.Image)
	} else if message.Image == "parametric" {
		// get image
		if !q.Features.IsUserInputFeatureActive() {
			log.Println("user input feature is disabled")
			return message.Text, fileState{}
		}

		input, err := q.Features.UserInput.GetUserInputCallback(telegramUserID)
		if err != nil {
			log.Println("get user input:", err)
			return message.Text, fileState{}
		}

		if message.ImageData.ArgumentType != "userInput" {
			log.Println(
				"unknown image argument type:",
				message.ImageData.ArgumentType,
			)
		}

		imageURL := fmt.Sprintf(
			message.ImageData.URLFormat,
			input,
		)

		imageData, err := swissknife.HttpGET(imageURL)
		if err != nil {
			log.Println("get image:", err)
			return message.Text, fileState{}
		}

		reader := bytes.NewReader(imageData)
		photo.File = tb.FromReader(reader)
	} else {
		// local image
		filePath := getFilePath(message.Image, filesRoot)
		if !swissknife.IsFileExists(filePath) {
			return message.Text, fileState{} // use plain text, when file not exists
		}

		photo.File = q.resCache.Get(message.Image)
		st.IsUsed = true
	}

	// add message text
	if message.Text != "" {
		photo.Caption = message.Text
	}
	return photo, st
}

func getTextMessage(message EventMessage) interface{} {
	return message.Text
}

// returns doc, is local
func (q *QueryHandler) getDocumentMessage(
	message EventMessage,
	filesRoot string,
) (interface{}, fileState) {
	st := fileState{
		IsUsed:        true,
		Type:          MessageTypeDocument,
		LocalFilePath: message.File.Path,
	}

	doc := &tb.Document{
		File:     q.resCache.Get(message.File.Path),
		FileName: message.File.Name,
	}
	if message.Text != "" {
		doc.Caption = message.Text
	}

	// add preview when available
	if message.File.PreviewImagePath != "" {
		previewPath := getFilePath(message.File.PreviewImagePath, filesRoot)

		if !swissknife.IsFileExists(previewPath) {
			log.Printf("file preview %q not exists, skip\n", previewPath)
		} else {
			doc.Thumbnail = &tb.Photo{
				File: q.resCache.Get(message.File.PreviewImagePath),
			}
		}
	}

	return doc, st
}

// returns doc, is local
func (q *QueryHandler) getAudioMessage(
	message EventMessage,
	filesRoot string,
) (interface{}, fileState) {
	st := fileState{
		IsUsed:        true,
		Type:          MessageTypeAudio,
		LocalFilePath: message.Audio.Path,
	}

	fileFullPath := getFilePath(
		message.Audio.Path,
		filesRoot,
	)
	if !swissknife.IsFileExists(fileFullPath) {
		log.Printf("audio file %q not found\n", fileFullPath)
		return getTextMessage(message), fileState{}
	}

	audio := &tb.Audio{
		File: q.resCache.Get(message.Audio.Path),
	}

	if message.Audio.Name != "" {
		audio.FileName = message.Audio.Name
	}
	if message.Audio.Duration > 0 {
		audio.Duration = message.Audio.Duration
	}
	if message.Text != "" {
		audio.Caption = message.Text
	}

	return audio, st
}

// returns doc, is local
func (q *QueryHandler) getVideoMessage(
	message EventMessage,
	filesRoot string,
) (interface{}, fileState) {
	st := fileState{
		IsUsed:        true,
		Type:          MessageTypeVideo,
		LocalFilePath: message.Image,
	}

	videoPath := getFilePath(
		message.Video.Path,
		filesRoot,
	)
	if !swissknife.IsFileExists(videoPath) {
		return "Failed to upload video for delivery. Try again later, sorry", fileState{}
	}

	video := &tb.Video{
		File:     tb.FromDisk(videoPath),
		Width:    message.Video.Width,
		Height:   message.Video.Height,
		FileName: "video.mp4",
	}
	if message.Text != "" {
		video.Caption = message.Text
	}

	if message.Video.PreviewImagePath != "" {
		previewPath := getFilePath(
			message.Video.PreviewImagePath,
			filesRoot,
		)

		if !swissknife.IsFileExists(previewPath) {
			log.Printf("file preview %q not exists, skip\n", previewPath)
		} else {
			video.Thumbnail = &tb.Photo{
				File:   q.resCache.Get(message.Video.PreviewImagePath),
				Width:  message.Video.Width,
				Height: message.Video.Height,
			}
		}
	}
	return video, st
}

func addUtmTags(baseURL string, tags UTMTags) (string, error) {
	if tags.Campaign == "" || tags.Source == "" {
		return baseURL, nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}

	q := u.Query()
	q.Set("utm_source", tags.Source)
	q.Set("utm_campaign", tags.Campaign)
	if tags.Content != "" {
		q.Set("utm_content", tags.Content)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// payload format: source_campaign_yclid
// example: dzen_start
// or: dzen_start_100
func FilterUserPayload(
	payloadRaw string,
) (UserPayload, error) {
	if payloadRaw == "" {
		return UserPayload{}, nil
	}

	if len(payloadRaw) > 168 {
		return UserPayload{}, errors.New("ignore heavy user payload")
	}

	if isBase64(payloadRaw) {
		return parseUTMBase64(payloadRaw)
	}

	return parseUTMSimple(payloadRaw)
}

func parseUTMBase64(payloadRaw string) (UserPayload, error) {
	payloadBytes, err := decodeBase64WithoutPadding(payloadRaw)
	if err != nil {
		return UserPayload{}, fmt.Errorf("base64: %w", err)
	}

	payloadData := string(payloadBytes)

	params, err := url.ParseQuery(payloadData)
	if err != nil {
		return UserPayload{}, fmt.Errorf("parse url data: %w", err)
	}

	return UserPayload{
		UTMSource:       params.Get("s"),
		UTMCampaign:     params.Get("c"),
		UTMContent:      params.Get("t"),
		BackLinkEventID: params.Get("b"),
		Yclid:           params.Get("y"),
	}, nil
}

func parseUTMSimple(payloadRaw string) (UserPayload, error) {
	parts := strings.Split(payloadRaw, "_")
	if len(parts) < 2 {
		return UserPayload{}, fmt.Errorf("invalid payload: %q", payloadRaw)
	}

	utmSource := LimitStringLen(parts[0], 24)
	utmCampaign := LimitStringLen(parts[1], 24)

	var yclid string
	if len(parts) > 2 {
		yclidRaw := LimitStringLen(parts[2], 64)
		if IsNumber(yclidRaw) {
			yclid = yclidRaw
		}
	}

	payload := UserPayload{
		UTMSource:   utmSource,
		UTMCampaign: utmCampaign,
		Yclid:       yclid,
	}

	if utmCampaign == "back" {
		payload.BackLinkEventID = utmSource
	}

	return payload, nil
}

// decodeBase64WithoutPadding декодирует строку Base64, которая может не содержать символов "=".
func decodeBase64WithoutPadding(encoded string) ([]byte, error) {
	// Вычисляем количество недостающих символов "="
	padding := len(encoded) % 4
	if padding > 0 {
		// Добавляем недостающие символы "="
		switch padding {
		case 1:
			return nil, fmt.Errorf(
				"not enought data in: %q",
				encoded,
			)
		case 2:
			encoded += "=="
		case 3:
			encoded += "="
		}
	}

	// Декодируем строку Base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return decoded, nil
}

// isBase64 проверяет, является ли строка корректной строкой Base64.
func isBase64(encoded string) bool {
	encoded = strings.ReplaceAll(encoded, "=", "")

	// Регулярное выражение для проверки символов Base64
	base64Regex := `^[A-Za-z0-9+/]*$`

	// Проверяем, соответствует ли строка регулярному выражению
	matched, err := regexp.MatchString(base64Regex, encoded)
	if err != nil {
		return false
	}

	// Проверяем длину строки: она должна быть кратна 4 или 2 или 3
	length := len(encoded)
	if length%4 == 0 || length%4 == 2 || length%4 == 3 {
		return matched
	}

	return false
}

// IsNumber проверяет, является ли строка числом (целым или дробным).
func IsNumber(s string) bool {
	// Проверяем, может ли строка быть преобразована в целое число
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}

	// Проверяем, может ли строка быть преобразована в число с плавающей запятой
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

func LimitStringLen(str string, maxLength int) string {
	if maxLength == 0 {
		return ""
	}

	if len(str) == maxLength || maxLength > len(str) {
		return str
	}

	return str[0:maxLength]
}

func findFileInMessage(m *tb.Message, st fileState) (tb.File, bool) {
	if m == nil {
		return tb.File{}, false
	}

	switch st.Type {
	default:
		return tb.File{}, false // skip
	case MessageTypeAudio:
		if m.Audio == nil {
			return tb.File{}, false
		}
		return m.Audio.File, true
	case MessageTypePhoto:
		if m.Audio == nil {
			return tb.File{}, false
		}
		return m.Photo.File, true
	case MessageTypeDocument:
		if m.Audio == nil {
			return tb.File{}, false
		}
		return m.Document.File, true
	case MessageTypeVideo:
		if m.Audio == nil {
			return tb.File{}, false
		}
		return m.Video.File, true
	}
}
