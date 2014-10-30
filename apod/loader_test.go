package apod

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDownload(t *testing.T) {
	a, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	_, err := a.loader.Download(testDateSeptember)
	assert.NoError(t, err)
	image := a.Config.fileName(testDateSeptember)
	i, err := os.Open(image)
	assert.NoError(t, err)
	info, err := i.Stat()
	assert.NoError(t, err)
	assert.Equal(t, 1375, info.Size(), "Wrong downloaded file size")
}

func TestLoadPeriod(t *testing.T) {
	a, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	err := a.loader.LoadPeriod(ADate("140925"), 5)
	assert.NoError(t, err)
	for _, dateS := range []string{"140924", "140923", "140921", "140920"} {
		file := a.Config.fileName(ADate(dateS))
		present, err := exists(file)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, present, "%s should exist", file)
	}
}
