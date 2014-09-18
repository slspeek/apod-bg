package apod

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func makeTestWallpapers(t testing.TB, s *Storage) {
	files := []string{"140120", "140121", "140122"}
	for _, file := range files {
		err := ioutil.WriteFile(s.fileName(file), []byte{}, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestFileName(t *testing.T) {
	apod := Storage{Config: &config{WallpaperDir: "foo"}}
	expected := filepath.Join("foo", "apod-img-140121")
	got := apod.fileName(testDateString)
	if expected != got {
		t.Fatalf("Expected: %v, got %v", expected, got)
	}
}

func TestDownloadedWallpapers(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, imageRoundTrip{})
	defer os.RemoveAll(testHome)
	makeTestWallpapers(t, a.storage)

	files, err := a.storage.DownloadedWallpapers()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatal("Expected 3 files")
	}
}

func TestIndexOf(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, imageRoundTrip{})
	defer os.RemoveAll(testHome)
	makeTestWallpapers(t, a.storage)
	i, err := a.storage.IndexOf("140121")
	if err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatalf("Expected 1, got %d", i)
	}
}
