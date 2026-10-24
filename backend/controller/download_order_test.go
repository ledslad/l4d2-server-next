package controller

import (
	"fmt"
	"testing"
	"time"
)

func TestDownloadHistoryNewestBeforePagination(t *testing.T) {
	d := NewDownloader()
	base := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	// More than two UI pages, deliberately stored out of timestamp order.
	for i := 0; i < 25; i++ {
		task := ownedDownloadTestTask("owner", DOWNLOAD_STATUS_COMPLETED)
		task.id = fmt.Sprintf("task-%02d", i)
		task.createdAt = base.Add(time.Duration((i*7)%25) * time.Minute)
		d.tasks = append(d.tasks, task)
	}
	foreign := ownedDownloadTestTask("other", DOWNLOAD_STATUS_COMPLETED)
	foreign.createdAt = base.Add(24 * time.Hour)
	d.tasks = append(d.tasks, foreign)
	originalFirst := d.tasks[0]
	listed := d.getTasksInfoForOwner("owner")
	if len(listed) != 25 {
		t.Fatalf("owner list has %d records, want 25", len(listed))
	}
	for offset := 0; offset < len(listed); offset += 10 {
		end := min(offset+10, len(listed))
		for i, item := range listed[offset:end] {
			want := base.Add(time.Duration(24-offset-i) * time.Minute)
			if got := item["createdAt"].(time.Time); !got.Equal(want) {
				t.Fatalf("page %d item %d time = %s, want %s", offset/10+1, i, got, want)
			}
		}
	}
	if d.tasks[0] != originalFirst {
		t.Fatal("listing reordered stored tasks, which would break index-based actions")
	}
}

func TestDownloadHistoryLegacyAndEqualTimesHaveStableFallback(t *testing.T) {
	d := NewDownloader()
	now := time.Now()
	for i, stamp := range []time.Time{{}, now, {}, now} {
		task := newDownloadConcurrencyTestTask(DOWNLOAD_STATUS_COMPLETED)
		task.id = fmt.Sprint(i)
		task.createdAt = stamp
		d.tasks = append(d.tasks, task)
	}
	for attempt := 0; attempt < 3; attempt++ {
		listed := d.GetTasksInfo()
		for i, want := range []string{"3", "1", "2", "0"} {
			if got := listed[i]["id"]; got != want {
				t.Fatalf("attempt %d item %d = %v, want %s", attempt, i, got, want)
			}
		}
	}
}
