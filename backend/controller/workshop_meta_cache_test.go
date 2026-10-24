package controller

import (
	"l4d2-manager-next/logic"
	"testing"
)

func TestParsedWorkshopMetadataCanBeRecoveredByDownloadURL(t *testing.T) {
	parsedWorkshopMetadata.Lock()
	parsedWorkshopMetadata.items = make(map[string]cachedWorkshopMeta)
	parsedWorkshopMetadata.Unlock()

	rememberParsedWorkshopMetadata(logic.LinkParseResult{
		SourceType: logic.LinkSourceWorkshop,
		Items: []logic.LinkParseItem{{
			ID:         "123734884",
			Title:      "Open Road",
			FileURL:    "https://cdn.example/map.vpk",
			PreviewURL: "https://cdn.example/preview.jpg",
		}},
	})
	meta := getParsedWorkshopMetadata("https://cdn.example/map.vpk")
	if meta == nil || meta.WorkshopID != "123734884" || meta.PreviewURL == "" {
		t.Fatalf("cached metadata = %+v", meta)
	}
}

func TestNonWorkshopParseResultIsNotCached(t *testing.T) {
	parsedWorkshopMetadata.Lock()
	parsedWorkshopMetadata.items = make(map[string]cachedWorkshopMeta)
	parsedWorkshopMetadata.Unlock()

	rememberParsedWorkshopMetadata(logic.LinkParseResult{
		SourceType: logic.LinkSourceQQFlashTransfer,
		Items:      []logic.LinkParseItem{{ID: "123734884", FileURL: "https://cdn.example/map.vpk"}},
	})
	if meta := getParsedWorkshopMetadata("https://cdn.example/map.vpk"); meta != nil {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}
