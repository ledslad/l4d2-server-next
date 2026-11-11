package logic

import (
	"encoding/json"
	"l4d2-manager-next/consts"
	"os"
	"path/filepath"
	"testing"
)

func TestPersistDeletedMapRemovalCreatesCompatibleRecord(t *testing.T) {
	oldAddons := consts.AddonsBasePath
	t.Cleanup(func() { consts.AddonsBasePath = oldAddons })
	consts.AddonsBasePath = t.TempDir()

	removal := DeletedMapRemoval{
		Filename:    "campaign.vpk",
		ID:          "campaign",
		Size:        "1.00MB",
		WorkshopID:  "123734884",
		WorkshopIDs: []string{"123734884"},
		Titles:      []string{"Campaign"},
		Maps:        []string{"c1m1_hotel"},
	}
	if err := removal.PersistDeletedMapRemoval("removed", "删除成功！"); err != nil {
		t.Fatalf("persist deleted metadata: %v", err)
	}

	paths, err := filepath.Glob(filepath.Join(consts.AddonsBasePath, "deleted-vpk-meta", "campaign.vpk.deleted-*.json"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("metadata paths = %v, err = %v", paths, err)
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("parse metadata: %v", err)
	}
	if record["status"] != "removed" || record["manager_response"] != "删除成功！" {
		t.Fatalf("record status/response = %#v", record)
	}
}

func TestPersistDeletedMapRemovalCompletesPendingRecordWithoutDuplicate(t *testing.T) {
	oldAddons := consts.AddonsBasePath
	t.Cleanup(func() { consts.AddonsBasePath = oldAddons })
	consts.AddonsBasePath = t.TempDir()
	metaDir := filepath.Join(consts.AddonsBasePath, "deleted-vpk-meta")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		t.Fatalf("create metadata directory: %v", err)
	}
	pending := `{"removed_at":"20261111-120000-123456","query":"123734884","id":"campaign","workshop_id":"123734884","workshop_ids":["123734884"],"filename":"campaign.vpk","size":"1.00MB","titles":["Campaign"],"maps":["c1m1_hotel"],"status":"pending","manager_response":""}`
	path := filepath.Join(metaDir, "campaign.vpk.deleted-20261111-120000-123456.json")
	if err := os.WriteFile(path, []byte(pending), 0644); err != nil {
		t.Fatalf("write pending metadata: %v", err)
	}

	removal := DeletedMapRemoval{Filename: "campaign.vpk", ID: "campaign", Size: "unknown"}
	if err := removal.PersistDeletedMapRemoval("removed", "删除成功！"); err != nil {
		t.Fatalf("complete pending metadata: %v", err)
	}
	paths, _ := filepath.Glob(filepath.Join(metaDir, "campaign.vpk.deleted-*.json"))
	if len(paths) != 1 {
		t.Fatalf("metadata paths = %v, want one", paths)
	}
	data, _ := os.ReadFile(paths[0])
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("parse completed metadata: %v", err)
	}
	if record["removed_at"] != "20261111-120000-123456" || record["query"] != "123734884" || record["status"] != "removed" {
		t.Fatalf("pending record was not preserved/completed: %#v", record)
	}
}
