package logic

import "testing"

func TestCleanWorkshopDownloadFilenameUsesItemID(t *testing.T) {
	item := WorkshopDownloadItem{
		PublishedFileID: "1234567890",
		Title:           "A Different Display Title",
		Filename:        "Campaign Name! (Final).vpk",
	}
	if got := cleanWorkshopDownloadFilename(item); got != "1234567890.vpk" {
		t.Fatalf("cleanWorkshopDownloadFilename() = %q, want %q", got, "1234567890.vpk")
	}
}

func TestCleanWorkshopDownloadFilenameDoesNotUseTitleFallback(t *testing.T) {
	item := WorkshopDownloadItem{
		PublishedFileID: "1234567890",
		Title:           "Display Title",
	}
	if got := cleanWorkshopDownloadFilename(item); got != "1234567890.vpk" {
		t.Fatalf("cleanWorkshopDownloadFilename() = %q", got)
	}
}

func TestCleanWorkshopDownloadFilenameHandlesMissingSourceIdentity(t *testing.T) {
	if got := cleanWorkshopDownloadFilename(WorkshopDownloadItem{}); got != "workshop.vpk" {
		t.Fatalf("cleanWorkshopDownloadFilename() = %q", got)
	}
}
