package apod

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func makeTestWallpapers(t testing.TB, c *config, files ...string) {
	for _, file := range files {
		err := ioutil.WriteFile(c.fileName(ADate(file)), []byte{}, 0644)
		assert.NoError(t, err)
	}
}

func TestFileName(t *testing.T) {
	c := config{WallpaperDir: "foo"}
	expected := filepath.Join("foo", "apod-img-140121")
	assert.Equal(t, expected, c.fileName(testDateString))
}

func TestDownloadedWallpapers(t *testing.T) {
	a, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	makeTestWallpapers(t, a.Config, "140120", "140121", "140122")

	files, err := a.storage.DownloadedWallpapers()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(files))
}

func TestIndexOfPresent(t *testing.T) {
	a, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	makeTestWallpapers(t, a.Config, "140120", "140121")
	i, err := a.storage.IndexOf("140121")
	assert.NoError(t, err)
	assert.Equal(t, 1, i)
}

func TestIndexOfAbsent(t *testing.T) {
	a, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	makeTestWallpapers(t, a.Config)
	_, err := a.storage.IndexOf("130101")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	assert.Equal(t, "130101 was not found", err.Error())
}
