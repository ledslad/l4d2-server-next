package logic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkshopMetaRoundTrip(t *testing.T) {
	dir := t.TempDir()
	vpkPath := filepath.Join(dir, "campaign.vpk")
	meta := WorkshopMeta{
		WorkshopID: "3571020978",
		Title:      "Campaign",
		PreviewURL: "https://example.com/preview.jpg",
	}
	if err := SaveWorkshopMetaForVPK(vpkPath, meta); err != nil {
		t.Fatalf("save metadata: %v", err)
	}
	loaded, err := LoadWorkshopMetaForVPK(vpkPath)
	if err != nil {
		t.Fatalf("load metadata: %v", err)
	}
	if loaded == nil || loaded.WorkshopID != meta.WorkshopID || loaded.Title != meta.Title {
		t.Fatalf("loaded metadata = %+v", loaded)
	}
	if loaded.DownloadedAt == "" {
		t.Fatal("downloaded_at was not populated")
	}
	if _, err := os.Stat(filepath.Join(dir, "campaign.meta")); err != nil {
		t.Fatalf("sidecar not created: %v", err)
	}
}

func TestLoadWorkshopMetaMissingAndInvalid(t *testing.T) {
	dir := t.TempDir()
	vpkPath := filepath.Join(dir, "campaign.vpk")
	loaded, err := LoadWorkshopMetaForVPK(vpkPath)
	if err != nil || loaded != nil {
		t.Fatalf("missing metadata = %+v, %v", loaded, err)
	}
	if err := os.WriteFile(MapMetaPath(vpkPath), []byte(`{"workshop_id":"invalid"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWorkshopMetaForVPK(vpkPath); err == nil {
		t.Fatal("invalid workshop ID should fail")
	}
}
