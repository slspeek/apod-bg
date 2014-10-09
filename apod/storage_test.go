package apod

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func makeTestWallpapers(t testing.TB, c *config, files ...string) {
	for _, file := range files {
		err := ioutil.WriteFile(c.fileName(ADate(file)), []byte{}, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestFileName(t *testing.T) {
	c := config{WallpaperDir: "foo"}
	expected := filepath.Join("foo", "apod-img-140121")
	got := c.fileName(testDateString)
	if expected != got {
		t.Fatalf("Expected: %v, got %v", expected, got)
	}
}

func TestDownloadedWallpapers(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, imageRoundTrip{})
	defer cleanUp(t, testHome)
	makeTestWallpapers(t, a.Config, "140120", "140121", "140122")

	files, err := a.storage.DownloadedWallpapers()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatal("Expected 3 files")
	}
}

func TestIndexOfPresent(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, imageRoundTrip{})
	defer cleanUp(t, testHome)
	makeTestWallpapers(t, a.Config, "140120", "140121")
	i, err := a.storage.IndexOf("140121")
	if err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatalf("Expected 1, got %d", i)
	}
}
func TestIndexOfAbsent(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, imageRoundTrip{})
	defer cleanUp(t, testHome)
	makeTestWallpapers(t, a.Config)
	_, err := a.storage.IndexOf("130101")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	if err.Error() != "130101 was not found" {
		t.Fatalf("Wrong error: %v", err.Error())
	}
}
