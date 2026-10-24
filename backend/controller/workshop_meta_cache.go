package controller

import (
	"l4d2-manager-next/logic"
	"strings"
	"sync"
	"time"
)

const parsedWorkshopMetaTTL = 30 * time.Minute

type cachedWorkshopMeta struct {
	meta      logic.WorkshopMeta
	expiresAt time.Time
}

var parsedWorkshopMetadata = struct {
	sync.Mutex
	items map[string]cachedWorkshopMeta
}{items: make(map[string]cachedWorkshopMeta)}

func rememberParsedWorkshopMetadata(result logic.LinkParseResult) {
	if result.SourceType != logic.LinkSourceWorkshop {
		return
	}
	now := time.Now()
	parsedWorkshopMetadata.Lock()
	defer parsedWorkshopMetadata.Unlock()
	for key, item := range parsedWorkshopMetadata.items {
		if now.After(item.expiresAt) {
			delete(parsedWorkshopMetadata.items, key)
		}
	}
	for _, item := range result.Items {
		url := strings.TrimSpace(item.FileURL)
		if url == "" || strings.TrimSpace(item.ID) == "" {
			continue
		}
		parsedWorkshopMetadata.items[url] = cachedWorkshopMeta{
			meta: logic.WorkshopMeta{
				WorkshopID:  item.ID,
				Title:       item.Title,
				Author:      item.Author,
				Description: item.Description,
				PreviewURL:  item.PreviewURL,
				FileURL:     item.FileURL,
				TimeUpdated: item.TimeUpdated,
			},
			expiresAt: now.Add(parsedWorkshopMetaTTL),
		}
	}
}

func getParsedWorkshopMetadata(downloadURL string) *logic.WorkshopMeta {
	downloadURL = strings.TrimSpace(downloadURL)
	parsedWorkshopMetadata.Lock()
	defer parsedWorkshopMetadata.Unlock()
	item, ok := parsedWorkshopMetadata.items[downloadURL]
	if !ok {
		return nil
	}
	if time.Now().After(item.expiresAt) {
		delete(parsedWorkshopMetadata.items, downloadURL)
		return nil
	}
	meta := item.meta
	return &meta
}
