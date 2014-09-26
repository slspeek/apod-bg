package apod

import (
	"os"
	"testing"
)

func TestDownload(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, testRoundTrip{})
	defer cleanUp(t, testHome)
	_, err := a.loader.Download(testDateString)
	if err != nil {
		t.Fatalf("could not load page: %v", err)
	}
	image := a.storage.fileName(testDateString)
	i, err := os.Open(image)
	if err != nil {
		t.Fatal(err)
	}
	info, err := i.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 2185 {
		t.Fatalf("Wrong size expected 2185, got: %d", info.Size())
	}
}
