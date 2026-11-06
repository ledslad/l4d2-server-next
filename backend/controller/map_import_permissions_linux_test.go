package controller

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestMapImportNormalizesPermissions(t *testing.T) {
	for _, archive := range []bool{false, true} {
		for _, mode := range []os.FileMode{0600, 0777, 0000} {
			t.Run(fmtPermissionCase(archive, mode), func(t *testing.T) {
				root, addons := setupMapImportTest(t)
				oldMask := syscall.Umask(0077)
				defer syscall.Umask(oldMask)
				payload := mapImportVPK(t, map[string][]byte{"missions/test.txt": []byte("test")})
				name := "permissions.vpk"
				source := filepath.Join(root, name)
				if archive {
					source = filepath.Join(root, "maps.zip")
					var data bytes.Buffer
					zw := zip.NewWriter(&data)
					header := &zip.FileHeader{Name: name, Method: zip.Store}
					header.SetMode(mode)
					entry, err := zw.CreateHeader(header)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := entry.Write(payload); err != nil {
						t.Fatal(err)
					}
					if err := zw.Close(); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(source, data.Bytes(), 0600); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(source, payload, 0600); err != nil {
						t.Fatal(err)
					}
					if err := os.Chmod(source, mode); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := ProcessFile(source); err != nil {
					t.Fatal(err)
				}
				destination := filepath.Join(addons, name)
				info, err := os.Stat(destination)
				if err != nil {
					t.Fatal(err)
				}
				if info.Mode().Perm() != 0644 {
					t.Fatalf("mode = %o; want 644", info.Mode().Perm())
				}
				got, err := os.ReadFile(destination)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, payload) {
					t.Fatal("VPK content changed")
				}
			})
		}
	}
}

func fmtPermissionCase(archive bool, mode os.FileMode) string {
	return fmt.Sprintf("archive=%t/mode=%03o", archive, mode)
}
