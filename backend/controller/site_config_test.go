package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLatestReleaseUsesConfiguredForkAndValidatesLink(t *testing.T) {
	link := "https://github.com/example/community-next/releases/tag/v2.0.0"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/example/community-next/releases/latest" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("must not forward manager credentials to GitHub")
		}
		fmt.Fprintf(w, `{"tag_name":"v2.0.0","html_url":%q}`, link)
	}))
	defer upstream.Close()
	release, err := fetchLatestRelease(context.Background(), upstream.Client(), upstream.URL, "example/community-next")
	if err != nil || release.TagName != "v2.0.0" {
		t.Fatalf("release=%#v err=%v", release, err)
	}
	link = "https://github.com/LaoYutang/l4d2-server-next/releases/tag/v2.0.0"
	if _, err := fetchLatestRelease(context.Background(), upstream.Client(), upstream.URL, "example/community-next"); err == nil {
		t.Fatal("accepted release from a different repository")
	}
}

func TestSiteConfigWriteRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/site/config/update", func(c *gin.Context) { c.Set("role", "guest"); SetSiteConfig(c) })
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/site/config/update", nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status=%d", w.Code)
	}
}
