package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitalConfiguration(t *testing.T) {
	c := exec.Command("go", "build", "-v")
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("apod-bg")
	ac := exec.Command("./apod-bg", "-config", "barewm")
	env := os.Environ()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testHome := filepath.Join(wd, "home-e2e")
	if err := os.MkdirAll(testHome, 0755); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(testHome)
	env = []string{fmt.Sprintf("HOME=%s", testHome)}
	ac.Env = env

	out, err = ac.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "apod-bg was successfully configured") {
		t.Fatalf("Expected success got: %s", string(out))
	}
}
