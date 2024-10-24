package tgfun

import (
	"fmt"
	"log"
	"sync"

	swissknife "github.com/Sagleft/swiss-knife"
	"gopkg.in/telebot.v3"
)

type Resource struct {
	FileUniqueID string `json:"fileUniqueID"`
	FileID       string `json:"fileID"`
	Size         int64  `json:"size"`
	Hash         string `json:"hash"`
}

type ResourcesCache struct {
	enabled bool
	root    string
	path    string
	data    sync.Map

	writeLocker sync.Mutex
	hashLocker  sync.Mutex
}

func NewResourceCache(cachePath string, pathRoot string) *ResourcesCache {
	if cachePath == "" {
		log.Println("cache path is not set. skip")
		return &ResourcesCache{enabled: false}
	}

	r := &ResourcesCache{
		enabled: true,
		root:    pathRoot,
		path:    cachePath,
	}

	if swissknife.IsFileExists(cachePath) {
		if err := r.load(); err != nil {
			r.enabled = false
			log.Printf(
				"load funnel cache from %q: %s\n",
				cachePath, err.Error(),
			)
			return r
		}
	}
	return r
}

func (r *ResourcesCache) load() error {
	var rawData map[string]interface{}
	if err := swissknife.ParseStructFromJSONFile(r.path, &rawData); err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	for k, v := range rawData {
		r.data.Store(k, v)
	}
	return nil
}

func (r *ResourcesCache) save() error {
	r.writeLocker.Lock()
	defer r.writeLocker.Unlock()

	var rawData map[string]interface{}

	r.data.Range(func(key, value any) bool {
		rawData[key.(string)] = value.(Resource)
		return true
	})

	if err := swissknife.SaveStructToJSONFileIndent(rawData, r.path); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}

func (r *ResourcesCache) Get(localFilePath string) telebot.File {
	filePath := getFilePath(localFilePath, r.root)
	if !r.enabled {
		return telebot.FromDisk(filePath)
	}

	res, isExists := r.data.Load(localFilePath)
	if !isExists {
		return telebot.FromDisk(filePath)
	}

	resData := res.(Resource)
	return telebot.File{
		FileID:   resData.FileID,
		UniqueID: resData.FileUniqueID,
		FileSize: resData.Size,
	}
}

func (r *ResourcesCache) Actualize(
	localFilePath string,
	fileData telebot.File,
) {
	if !r.IsNeedUpdate(localFilePath, fileData) {
		return
	}

	if err := r.Update(localFilePath, fileData); err != nil {
		log.Printf(
			"failed to update res %q: %s\n",
			localFilePath, err.Error(),
		)
	}
}

func (r *ResourcesCache) IsNeedUpdate(
	localFilePath string,
	fileData telebot.File,
) bool {
	if !r.enabled {
		return false
	}

	// load
	var resData Resource
	resRaw, _ := r.data.LoadOrStore(localFilePath, Resource{})
	resData = resRaw.(Resource)

	// check
	r.hashLocker.Lock()
	defer r.hashLocker.Unlock()

	filePath := getFilePath(localFilePath, r.root)
	fileHash := r.getActualHash(filePath)
	return resData.Hash != fileHash && fileHash != ""
}

func (r *ResourcesCache) getActualHash(filePath string) string {
	if !swissknife.IsFileExists(filePath) {
		return ""
	}
	fileBytes, err := swissknife.ReadFileToBytes(filePath)
	if err != nil {
		log.Printf("res cache: read %q: %s\n", filePath, err.Error())
		return ""
	}

	// TODO: можно сделать и кеш MD5-хешей,
	// чтобы кеширование происходило только при запуске
	return swissknife.MD5(fileBytes)
}

func (r *ResourcesCache) Update(
	localFilePath string,
	fileData telebot.File,
) error {
	if !r.enabled {
		return nil
	}

	// load
	var resData Resource
	resRaw, isExists := r.data.Load(localFilePath)
	if isExists {
		resData = resRaw.(Resource)
	}

	// update
	resData.FileID = fileData.FileID
	resData.FileUniqueID = fileData.UniqueID
	resData.Size = fileData.FileSize
	resData.Hash = r.getActualHash(getFilePath(localFilePath, r.root))

	// save
	r.data.Store(localFilePath, resData)
	if err := r.save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	return nil
}
