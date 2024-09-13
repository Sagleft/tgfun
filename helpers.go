package tgfun

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	swissknife "github.com/Sagleft/swiss-knife"
	"github.com/microcosm-cc/bluemonday"
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

func getPhotoMessage(message EventMessage, filesRoot string) interface{} {
	photo := &tb.Photo{}
	if strings.Contains(message.Image, "http") {
		photo.File = tb.FromURL(message.Image)
	} else {
		filePath := getFilePath(message.Image, filesRoot)
		if !swissknife.IsFileExists(filePath) {
			return message.Text // use plain text, when file not exists
		}

		photo.File = tb.FromDisk(filePath)
	}
	if message.Text != "" {
		photo.Caption = message.Text
	}
	return photo
}

func getTextMessage(message EventMessage) interface{} {
	return message.Text
}

func getDocumentMessage(message EventMessage, filesRoot string) interface{} {
	docPath := getFilePath(
		message.File.Path,
		filesRoot,
	)
	if !swissknife.IsFileExists(docPath) {
		// when file not found
		return "Failed to upload file for delivery. Try again later, sorry"
	}

	doc := &tb.Document{
		File:     tb.FromDisk(docPath),
		FileName: message.File.Name,
	}
	if message.Text != "" {
		doc.Caption = message.Text
	}

	// add preview when available
	if message.File.PreviewImagePath != "" {
		previewPath := getFilePath(
			message.File.PreviewImagePath,
			filesRoot,
		)
		if !swissknife.IsFileExists(previewPath) {
			log.Printf("file preview %q not exists, skip\n", previewPath)
		} else {
			doc.Thumbnail = &tb.Photo{
				File: tb.FromDisk(previewPath),
			}
		}
	}

	return doc
}

func getAudioMessage(message EventMessage, filesRoot string) interface{} {
	filePath := getFilePath(
		message.Audio.Path,
		filesRoot,
	)
	if !swissknife.IsFileExists(filePath) {
		log.Printf("audio file %q not found\n", message.Audio.Path)
		return getTextMessage(message)
	}

	audio := &tb.Audio{
		File: tb.FromDisk(filePath),
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

	return audio
}

func getVideoMessage(message EventMessage, filesRoot string) interface{} {
	videoPath := getFilePath(
		message.Video.Path,
		filesRoot,
	)
	if !swissknife.IsFileExists(videoPath) {
		return "Failed to upload video for delivery. Try again later, sorry"
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
				File:   tb.FromDisk(previewPath),
				Width:  message.Video.Width,
				Height: message.Video.Height,
			}
		}
	}

	return video
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

func parseUTMTags(payload string, sanitizer *bluemonday.Policy) UTMTags {
	if sanitizer == nil {
		log.Println("parse utm tags: sanitizer is not set")
		return UTMTags{}
	}

	payloadParts := strings.Split(payload, "_")
	if len(payloadParts) < 2 {
		return UTMTags{}
	}

	return UTMTags{
		Source:   sanitizer.Sanitize(payloadParts[0]),
		Campaign: sanitizer.Sanitize(payloadParts[1]),
	}
}
