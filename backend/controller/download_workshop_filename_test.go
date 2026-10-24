package controller

import (
	"bytes"
	"l4d2-manager-next/logic"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkshopDownloadsOverrideCollidingCallerFilenames(t *testing.T) {
	semaphore := make(chan struct{}) // Block network access for these queued tasks.
	firstMeta := &logic.WorkshopMeta{WorkshopID: "3460058757", Title: "Campaign", PreviewURL: "https://example.com/cover.jpg"}
	secondMeta := &logic.WorkshopMeta{WorkshopID: "1586823798", Title: "Campaign resources"}
	first := newDownloadTask("https://example.invalid/one", "same-name.vpk", "", firstMeta, semaphore)
	second := newDownloadTask("https://example.invalid/two", "same-name.vpk", "", secondMeta, semaphore)
	t.Cleanup(first.Cancel)
	t.Cleanup(second.Cancel)
	if got := first.GetFilename(); got != "3460058757.vpk" {
		t.Fatalf("first filename = %q", got)
	}
	if got := second.GetFilename(); got != "1586823798.vpk" {
		t.Fatalf("second filename = %q", got)
	}
	if first.workshopMeta.Title != firstMeta.Title || first.workshopMeta.PreviewURL != firstMeta.PreviewURL {
		t.Fatal("renaming must preserve Workshop metadata")
	}
}

func TestSameNamedWorkshopVPKsInstallSeparatelyWithMetadata(t *testing.T) {
	setupDownloadConfigControllerTest(t)
	if err := logic.SetVPKTrimEnable(false); err != nil {
		t.Fatal(err)
	}
	_, addonsPath := setupRemoveTestPaths(t)
	// A valid VPK v1 with an empty directory. Use different trailing data to
	// detect accidental cross-task writes, not just different task labels.
	header := []byte{0x34, 0x12, 0xaa, 0x55, 1, 0, 0, 0, 1, 0, 0, 0, 0}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="same-name.vpk"`)
		w.Write(append(append([]byte(nil), header...), []byte(r.URL.Path)...))
	}))
	defer server.Close()
	first := newDownloadTask(server.URL+"/one", "same-name.vpk", "", &logic.WorkshopMeta{WorkshopID: "3460058757", Title: "Campaign"}, make(chan struct{}, 3))
	second := newDownloadTask(server.URL+"/two", "same-name.vpk", "", &logic.WorkshopMeta{WorkshopID: "1586823798", Title: "Resources"}, first.semaphore)
	t.Cleanup(first.Cancel)
	t.Cleanup(second.Cancel)
	for _, test := range []struct {
		task  *downloadTask
		id    string
		path  string
		title string
	}{
		{first, "3460058757", "/one", "Campaign"},
		{second, "1586823798", "/two", "Resources"},
	} {
		select {
		case <-test.task.done:
		case <-time.After(5 * time.Second):
			t.Fatal("download did not finish")
		}
		if test.task.GetStatus() != DOWNLOAD_STATUS_COMPLETED {
			t.Fatalf("download %s failed: %s", test.id, test.task.GetMessage())
		}
		path := filepath.Join(addonsPath, test.id+".vpk")
		data, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(data, append(append([]byte(nil), header...), []byte(test.path)...)) {
			t.Fatalf("VPK %s lost its own contents: %v", test.id, err)
		}
		meta, err := logic.LoadWorkshopMetaForVPK(path)
		if err != nil || meta == nil || meta.WorkshopID != test.id || meta.Title != test.title {
			t.Fatalf("VPK %s metadata = %+v, err = %v", test.id, meta, err)
		}
	}
	if _, err := os.Stat(filepath.Join(addonsPath, "same-name.vpk")); !os.IsNotExist(err) {
		t.Fatal("author filename was used instead of Workshop ID")
	}
}

func TestWorkshopMetadataSurvivesOwnedTaskRestart(t *testing.T) {
	store := setupAuthControllerStore(t)
	owner := createMapDownloadCode(t, store)
	d := setupOwnedDownloaderTest(t)
	meta := &logic.WorkshopMeta{WorkshopID: "3460058757", Title: "Campaign"}
	d.addTaskWithMetadataForOwner("https://example.invalid/file", "same-name.vpk", "", meta, owner.ID)
	id := d.tasks[0].id
	if !d.RestartTask(0) {
		t.Fatal("could not restart task")
	}
	if d.tasks[0].id != id || d.tasks[0].authCodeID != owner.ID || d.tasks[0].GetFilename() != "3460058757.vpk" || d.tasks[0].workshopMeta != meta {
		t.Fatal("restart lost owner, identity, ID filename or Workshop metadata")
	}
}

func TestOrdinaryDownloadsKeepCallerFilename(t *testing.T) {
	task := newDownloadTask("https://example.invalid/file", "ordinary-map.vpk", "", nil, make(chan struct{}))
	t.Cleanup(task.Cancel)
	if got := task.GetFilename(); got != "ordinary-map.vpk" {
		t.Fatalf("ordinary filename = %q", got)
	}
}
