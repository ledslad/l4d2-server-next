package controller

import (
	"l4d2-manager-next/consts"
	"l4d2-manager-next/logic"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestClearRemovesVPKAndWorkshopMeta(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, addonsPath := setupRemoveTestPaths(t)
	for _, name := range []string{"one.vpk", "two.vpk"} {
		vpkPath := filepath.Join(addonsPath, name)
		if err := os.WriteFile(vpkPath, []byte("map"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(logic.MapMetaPath(vpkPath), []byte(`{"workshop_id":"123734884"}`), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(consts.MapListFilePath, []byte("one.vpk\ntwo.vpk\n"), 0644); err != nil {
		t.Fatal(err)
	}
	c, recorder := newFormTestContext("/clear", nil)
	Clear(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", recorder.Code, recorder.Body.String())
	}
	for _, name := range []string{"one.vpk", "one.meta", "two.vpk", "two.meta"} {
		if _, err := os.Stat(filepath.Join(addonsPath, name)); !os.IsNotExist(err) {
			t.Fatalf("%s still exists: %v", name, err)
		}
	}
	data, err := os.ReadFile(consts.MapListFilePath)
	if err != nil || len(data) != 0 {
		t.Fatalf("maplist = %q, err = %v", string(data), err)
	}
}

func TestRenameMovesWorkshopMeta(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, addonsPath := setupRemoveTestPaths(t)
	oldPath := filepath.Join(addonsPath, "old.vpk")
	newPath := filepath.Join(addonsPath, "new.vpk")
	if err := os.WriteFile(oldPath, []byte("map"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logic.MapMetaPath(oldPath), []byte(`{"workshop_id":"123734884"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(consts.MapListFilePath, []byte("old.vpk\n"), 0644); err != nil {
		t.Fatal(err)
	}
	c, recorder := newFormTestContext("/rename", url.Values{"oldName": {"old.vpk"}, "newName": {"new.vpk"}})
	RenameMap(c)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", recorder.Code, recorder.Body.String())
	}
	for _, path := range []string{newPath, logic.MapMetaPath(newPath)} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("renamed file missing: %s: %v", path, err)
		}
	}
	for _, path := range []string{oldPath, logic.MapMetaPath(oldPath)} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("old file still exists: %s: %v", path, err)
		}
	}
}

func TestRenameRejectsStaleTargetMetaWithoutMovingVPK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, addonsPath := setupRemoveTestPaths(t)
	oldPath := filepath.Join(addonsPath, "old.vpk")
	newPath := filepath.Join(addonsPath, "new.vpk")
	if err := os.WriteFile(oldPath, []byte("map"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logic.MapMetaPath(newPath), []byte(`{"workshop_id":"123734884"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(consts.MapListFilePath, []byte("old.vpk\n"), 0644); err != nil {
		t.Fatal(err)
	}
	c, recorder := newFormTestContext("/rename", url.Values{"oldName": {"old.vpk"}, "newName": {"new.vpk"}})
	RenameMap(c)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %q", recorder.Code, recorder.Body.String())
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("old VPK was moved: %v", err)
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Fatalf("new VPK unexpectedly exists: %v", err)
	}
}

func TestRollbackInstalledMapsRemovesFilesAndRecords(t *testing.T) {
	_, addonsPath := setupRemoveTestPaths(t)
	vpkPath := filepath.Join(addonsPath, "failed.vpk")
	if err := os.WriteFile(vpkPath, []byte("map"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logic.MapMetaPath(vpkPath), []byte(`{"workshop_id":"123734884"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(consts.MapListFilePath, []byte("failed.vpk\nother.vpk\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := rollbackInstalledMaps([]string{"failed.vpk"}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{vpkPath, logic.MapMetaPath(vpkPath)} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("rolled back file still exists: %s: %v", path, err)
		}
	}
	data, err := os.ReadFile(consts.MapListFilePath)
	if err != nil || strings.Contains(string(data), "failed.vpk") || !strings.Contains(string(data), "other.vpk") {
		t.Fatalf("maplist = %q, err = %v", string(data), err)
	}
}
