package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestE2eInitalConfiguration(t *testing.T) {
	c := exec.Command("go", "build", "-v")
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("apod-bg")
	ac := exec.Command("./apod-bg", "-nonotify", "-config", "barewm")
	testHome := setupTestHome(t)
	defer os.RemoveAll(testHome)
	ac.Env = []string{fmt.Sprintf("HOME=%s", testHome)}

	out, err = ac.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "apod-bg was successfully configured") {
		t.Fatalf("Expected success got: %s", string(out))
	}
}
