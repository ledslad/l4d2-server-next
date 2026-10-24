package controller

import "testing"

func TestSanitizeFilenamePreservesSourceName(t *testing.T) {
	const source = "Campaign Name! (Final).VPK"
	if got := sanitizeFilename(source); got != source {
		t.Fatalf("sanitizeFilename(%q) = %q, want original name", source, got)
	}
}

func TestSanitizeFilenameReplacesOnlyUnsafeCharacters(t *testing.T) {
	const source = `folder/Map: Redux?.vpk`
	if got := sanitizeFilename(source); got != "Map_ Redux_.vpk" {
		t.Fatalf("sanitizeFilename(%q) = %q", source, got)
	}
}
