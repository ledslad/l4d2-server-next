package logic

import "testing"

func TestWorkshopItemsWithSameVPKNameStaySeparate(t *testing.T) {
	root := WorkshopDownloadItem{PublishedFileID: "3460058757", Filename: "campaign.vpk", Title: "Campaign"}
	dependency := WorkshopDownloadItem{PublishedFileID: "1586823798", Filename: "campaign.vpk", Title: "Campaign resources"}
	if got := cleanWorkshopDownloadFilename(root); got != "3460058757.vpk" {
		t.Fatalf("root filename = %q", got)
	}
	if got := cleanWorkshopDownloadFilename(dependency); got != "1586823798.vpk" {
		t.Fatalf("dependency must use its own ID, filename = %q", got)
	}
}

func TestWorkshopFilenameRejectsUnsafeIDs(t *testing.T) {
	for _, id := range []string{"", "../123734884", "123734884/other", "not-an-id", "123", "18446744073709551616"} {
		if got := WorkshopVPKFilename(id); got != "" {
			t.Fatalf("unsafe ID %q produced filename %q", id, got)
		}
	}
}
