package apod

import (
	"fmt"
	"github.com/haklop/gnotifier"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var testDate = time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)

const testDateString = "140121"
const testDateSeptember = "140924"
const testDateYoutube = "140922"
const configJSON = `{"WallpaperDir":"bar"}`
const setScriptSuccess = `#!/bin/bash
exit 0
`
const setScriptFailure = `#!/bin/bash
echo Something went wrong
echo Fault 1>&2
exit 5
`

func unsetNoseedFlag() {
	falseB := false
	noseed = &falseB
}

func setDateFlag(date string) {
	dateFlag = &date
	return
}

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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	t.Logf("%s was cleanly removed", dir)
}

func makeStateFile(t testing.TB, datecode, options string) {
	s := State{DateCode: ADate(datecode), Options: options}
	err := store(s)
	assert.NoError(t, err)
}

func frontendForTest(t *testing.T) (*Frontend, string) {
	recorder := gnotifier.NewTestRecorder()
	f := NewFrontend(nullLogger{}, Notifier{recorder.Notification})
	setDateFlag("140921")
	f.APOD.Site = testAPODSite
	return f, setupTestHome(t)
}

func frontendForTestConfigured(t *testing.T) (*Frontend, string) {
	f, testHome := frontendForTest(t)
	err := f.Configure("barewm")
	assert.NoError(t, err)
	err = f.Loadconfig()
	assert.NoError(t, err)
	assert.NoError(t, writeWallpaperScript(setScriptSuccess))
	return f, testHome
}

func TestSeed(t *testing.T) {
	f, testHome := frontendForTest(t)
	defer cleanUp(t, testHome)
	assert.NoError(t, writeWallpaperScript(setScriptSuccess))
	unsetNoseedFlag()
	err := f.Configure("barewm")
	assert.NoError(t, err)
}

func TestSeedYoutube(t *testing.T) {
	f, testHome := frontendForTest(t)
	defer cleanUp(t, testHome)
	assert.NoError(t, writeWallpaperScript(setScriptSuccess))
	setDateFlag(testDateYoutube)
	unsetNoseedFlag()
	err := f.Configure("barewm")
	assert.NoError(t, err)
	s, err := f.State()
	assert.NoError(t, err)
	assert.Equal(t, "140921", s.DateCode.String())
}

func TestWriteAutostart(t *testing.T) {
	f, testHome := frontendForTest(t)
	defer cleanUp(t, testHome)
	assert.NoError(t, f.writeAutostart())
	being, err := exists(filepath.Join(testHome, ".config", "autostart", "apod-bg.desktop"))
	assert.NoError(t, err)
	assert.True(t, being)
}

func TestRemoveAutostart(t *testing.T) {
	f, testHome := frontendForTest(t)
	defer cleanUp(t, testHome)
	assert.NoError(t, f.writeAutostart())
	assert.NoError(t, f.removeAutostart())
	being, err := exists(filepath.Join(testHome, ".config", "autostart", "apod-bg.desktop"))
	assert.NoError(t, err)
	assert.False(t, being)
}

func TestJumpAtBeginBackward(t *testing.T) {
	f, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	makeStateFile(t, "140120", "fit")
	makeTestWallpapers(t, f.Config, "140120")
	err := f.Jump(-1)
	assert.Equal(t, "Begin reached", err.Error(), "Wrong error message")
}

func TestJumpAtBeginForward(t *testing.T) {
	f, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	makeStateFile(t, "140120", "fit")
	makeTestWallpapers(t, f.Config, "140120")
	err := f.Jump(1)
	assert.Equal(t, "End reached", err.Error(), "Wrong error message")
}

func TestState(t *testing.T) {
	APOD, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	makeStateFile(t, "140121", "fit")
	rv, err := APOD.State()
	assert.NoError(t, err, "Error during call to State")
	assert.Equal(t, testDateString, rv.DateCode)
}

func RunConfiguration(t *testing.T, cfg string, expected string) {
	f, testHome := frontendForTest(t)
	defer cleanUp(t, testHome)
	f.Configure(cfg)
	script := wallpaperSetScript()
	bs, err := ioutil.ReadFile(script)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(bs))
}

func TestConfiguration(t *testing.T) {
	for _, cfg := range [][]string{[]string{"barewm", setScriptBareWM},
		[]string{"gnome", setScriptGNOME}, []string{"lxde", setScriptLXDE}} {
		RunConfiguration(t, cfg[0], cfg[1])
	}
}

func TestLoadconfigNonExistent(t *testing.T) {
	f, testHome := frontendForTest(t)
	defer cleanUp(t, testHome)

	err := f.Loadconfig()
	if err.Error() != configNotFound {
		t.Fatalf("Expected: %v got: %v", configNotFound, err)
	}
}

func TestLoadconfigExistent(t *testing.T) {
	APOD, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)

	err := APOD.Loadconfig()
	assert.NoError(t, err)
}

func TestToday(t *testing.T) {
	front := Frontend{}
	setDateFlag(testDateString)
	var ad ADate
	ad = front.Today()
	assert.Equal(t, ad.String(), testDateString)
}

func TestRandomArchiveEmpty(t *testing.T) {
	front, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	writeWallpaperScript(setScriptSuccess)
	err := front.RandomArchive()
	assert.Equal(t, "No backgrounds downloaded yet", err.Error())
}

func TestRandomArchiveSuccessTwoWallpapers(t *testing.T) {
	f, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	writeWallpaperScript(setScriptSuccess)
	makeTestWallpapers(t, f.Config, "140119", "140120")
	err := f.RandomArchive()
	assert.NoError(t, err)
}

func TestRandomArchiveFailureOneWallpaper(t *testing.T) {
	f, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	writeWallpaperScript(setScriptFailure)
	makeTestWallpapers(t, f.Config, "140119")
	err := f.RandomArchive()
	assert.Equal(t, "Error running Wallpaper-Set-Script: exit status 5. Output: Something went wrong\nFault\n", err.Error())
}

func TestRandomArchiveSuccessOneWallpaper(t *testing.T) {
	f, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	writeWallpaperScript(setScriptSuccess)
	makeTestWallpapers(t, f.Config, "140119")
	err := f.RandomArchive()
	assert.NoError(t, err)
}

func TestSetWallpaperSuccess(t *testing.T) {
	front, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	writeWallpaperScript(setScriptSuccess)
	err := front.SetWallpaper(State{DateCode: testDateString})
	assert.NoError(t, err)
}

func TestSetWallpaperFailure(t *testing.T) {
	front, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	writeWallpaperScript(setScriptFailure)
	err := front.SetWallpaper(State{DateCode: testDateString})
	assert.Equal(t, "Error running Wallpaper-Set-Script: exit status 5. Output: Something went wrong\nFault\n", err.Error())
}

func TestConfigurationE2e(t *testing.T) {
	testHome := setupTestHome(t)
	defer cleanUp(t, testHome)
	resetFlags()

	cfg := "barewm"
	configFlag = &cfg
	assert.NoError(t, Execute())
}

func TestJumpWithStateE2e(t *testing.T) {
	f, testHome := frontendForTestConfigured(t)
	defer cleanUp(t, testHome)
	makeStateFile(t, "140120", "fit")
	makeTestWallpapers(t, f.Config, "140119", "140120")
	writeWallpaperScript(setScriptSuccess)
	resetFlags()

	j := -1
	jump = &j
	assert.NoError(t, Execute())
}
