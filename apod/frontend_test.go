package apod

import (
	"fmt"
	"github.com/101loops/clock"
	"github.com/haklop/gnotifier"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testDateString = "140121"
const testDateSeptember = "140924"
const configJSON = `{"WallpaperDir":"bar"}`
const setScriptSuccess = `#!/bin/bash
exit 0
`
const setScriptFailure = `#!/bin/bash
echo Something went wrong
echo Fault 1>&2
exit 5
`

func resetFlags() {

	zero := 0
	trueB := true
	falseB := false
	emptyS := ""

	info = &falseB
	login = &falseB
	logFileFlag = &emptyS
	days = &zero
	jump = &zero
	configFlag = &emptyS
	apodFlag = &falseB
	mode = &falseB
	nonotify = &trueB
	noseed = &trueB
	randomFlag = &falseB
}

type nullLogger struct{}

func (n nullLogger) Printf(f string, i ...interface{}) {
}

func setupTestHome(t testing.TB) string {
	resetFlags()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testHome := filepath.Join(wd, fmt.Sprintf("test-home-%d", rand.Int63()))
	if err := os.MkdirAll(testHome, 0755); err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", testHome)
	t.Logf("%s CREATED", testHome)
	return testHome
}

func cleanUp(t testing.TB, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s was cleanly removed", dir)
}

func makeStateFile(t testing.TB, datecode, options string) {
	s := State{DateCode: ADate(datecode), Options: options}
	err := store(s)
	if err != nil {
		t.Fatal(err)
	}
}

func frontendForTest(t *testing.T, tripper http.RoundTripper) (*Frontend, string) {
	recorder := gnotifier.NewTestRecorder()
	f := NewFrontend(nullLogger{}, Notifier{recorder.Notification})
	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	m := clock.NewMock()
	m.Set(t0)
	f.Clock = m
	f.APOD.Client = &http.Client{Transport: tripper}

	return f, setupTestHome(t)
}

func frontendForTestConfigured(t *testing.T, tripper http.RoundTripper) (*Frontend, string) {
	f, testHome := frontendForTest(t, tripper)
	err := f.Configure("barewm")
	if err != nil {
		t.Fatal(err)
	}
	err = f.Loadconfig()
	if err != nil {
		t.Fatal(err)
	}
	return f, testHome
}

func TestJumpAtBeginBackward(t *testing.T) {
	f, testHome := frontendForTestConfigured(t, nil)
	defer cleanUp(t, testHome)
	makeStateFile(t, "140120", "fit")
	makeTestWallpapers(t, f.Config, "140120")
	writeWallpaperScript(setScriptSuccess)
	err := f.Jump(-1)
	if err.Error() != "Begin reached" {
		t.Fatalf("Wrong error: %v", err)
	}
}

func TestJumpAtBeginForward(t *testing.T) {
	f, testHome := frontendForTestConfigured(t, nil)
	defer cleanUp(t, testHome)
	makeStateFile(t, "140120", "fit")
	makeTestWallpapers(t, f.Config, "140120")
	writeWallpaperScript(setScriptSuccess)
	err := f.Jump(1)
	if err.Error() != "End reached" {
		t.Fatalf("Wrong error: %v", err)
	}
}

func TestState(t *testing.T) {
	APOD, testHome := frontendForTestConfigured(t, nil)
	defer cleanUp(t, testHome)
	makeStateFile(t, "140121", "fit")
	rv, err := APOD.State()
	if err != nil {
		t.Fatalf("Error during call to State: %v\n", err)
	}
	if rv.DateCode != testDateString {
		t.Errorf("Expected 140121, got %v", rv)
	}
}

func RunConfiguration(t *testing.T, cfg string, expected string) {
	f, testHome := frontendForTest(t, nil)
	defer cleanUp(t, testHome)
	f.Configure(cfg)
	script := wallpaperSetScript()
	bs, err := ioutil.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != expected {
		t.Fatalf("Expected %s, got: %s", setScriptBareWM, string(bs))
	}
}

func TestConfiguration(t *testing.T) {
	for _, cfg := range [][]string{[]string{"barewm", setScriptBareWM},
		[]string{"gnome", setScriptGNOME}, []string{"lxde", setScriptLXDE}} {
		RunConfiguration(t, cfg[0], cfg[1])
	}
}

func TestLoadconfigNonExistent(t *testing.T) {
	f, testHome := frontendForTest(t, nil)
	defer cleanUp(t, testHome)

	err := f.Loadconfig()
	if err.Error() != configNotFound {
		t.Fatalf("Expected: %v got: %v", configNotFound, err)
	}
}

func TestLoadconfigExistent(t *testing.T) {
	APOD, testHome := frontendForTestConfigured(t, nil)
	defer cleanUp(t, testHome)

	err := APOD.Loadconfig()
	if err != nil {
		t.Fatal(err)
	}
}

func TestToday(t *testing.T) {
	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	m := clock.NewMock()
	m.Set(t0)

	front := Frontend{Clock: m}
	var ad ADate
	ad = front.Today()
	assert.Equal(t, ad.String(), testDateString)
}

func TestSetWallpaperSuccess(t *testing.T) {
	front, testHome := frontendForTestConfigured(t, testRoundTrip{})
	defer cleanUp(t, testHome)
	writeWallpaperScript(setScriptSuccess)
	err := front.SetWallpaper(State{DateCode: testDateString})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetWallpaperFailure(t *testing.T) {
	front, testHome := frontendForTestConfigured(t, testRoundTrip{})
	defer cleanUp(t, testHome)
	writeWallpaperScript(setScriptFailure)
	err := front.SetWallpaper(State{DateCode: testDateString})
	if err.Error() != "Script error: exit status 5. Output: Something went wrong\nFault\n" {
		t.Fatalf("Wrong error: %v", err)
	}
}

func TestConfigurationE2e(t *testing.T) {
	testHome := setupTestHome(t)
	defer cleanUp(t, testHome)
	resetFlags()

	cfg := "barewm"
	configFlag = &cfg
	if err := Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestJumpWithStateE2e(t *testing.T) {
	f, testHome := frontendForTestConfigured(t, nil)
	defer cleanUp(t, testHome)
	makeStateFile(t, "140120", "fit")
	makeTestWallpapers(t, f.Config, "140119", "140120")
	writeWallpaperScript(setScriptSuccess)
	resetFlags()

	j := -1
	jump = &j
	if err := Execute(); err != nil {
		t.Fatal(err)
	}
}
