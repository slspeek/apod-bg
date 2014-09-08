package main

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestHome(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testHome := filepath.Join(wd, "test-home")
	if err := os.MkdirAll(testHome, 0755); err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", testHome)
	return testHome
}

func TestConfigurationE2e(t *testing.T) {
	testHome := setupTestHome(t)
	defer os.RemoveAll(testHome)
	cfg := "barewm"
	config = &cfg
	execute()
}
