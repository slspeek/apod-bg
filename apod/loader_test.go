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
