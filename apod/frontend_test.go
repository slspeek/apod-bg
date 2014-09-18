package apod

import (
	"github.com/101loops/clock"
	"github.com/haklop/gnotifier"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type nullLogger struct{}

func (n nullLogger) Printf(f string, i ...interface{}) {
}

func setupTestHome(t testing.TB) string {
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

func makeStateFile(t testing.TB) {
	err := ioutil.WriteFile(stateFile(), []byte(`{"DateCode":"140121","Options":"fit"}`), 0644)
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

func TestJump(t *testing.T) {
	f, testHome := frontendForTestConfigured(t, imageRoundTrip{})
	defer os.RemoveAll(testHome)
	makeTestWallpapers(t, f.storage)
	writeWallpaperScript(setScriptSuccess)
	err := f.Jump(-1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestState(t *testing.T) {
	APOD, testHome := frontendForTestConfigured(t, nil)
	defer os.RemoveAll(testHome)
	makeStateFile(t)
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
	defer os.RemoveAll(testHome)
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
	defer os.RemoveAll(testHome)

	err := f.Loadconfig()
	if err.Error() != configNotFound {
		t.Fatalf("Expected: %v got: %v", configNotFound, err)
	}
}

func TestLoadconfigExistent(t *testing.T) {
	APOD, testHome := frontendForTestConfigured(t, nil)
	defer os.RemoveAll(testHome)

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
	if front.Today() != testDateString {
		t.Errorf("Expected %v, got %v", testDateString, front.Today())
	}
}

func TestSetWallpaperSuccess(t *testing.T) {
	front, testHome := frontendForTestConfigured(t, testRoundTrip{})
	defer os.RemoveAll(testHome)
	writeWallpaperScript(setScriptSuccess)
	err := front.SetWallpaper(State{DateCode: testDateString})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetWallpaperFailure(t *testing.T) {
	front, testHome := frontendForTestConfigured(t, testRoundTrip{})
	defer os.RemoveAll(testHome)
	writeWallpaperScript(setScriptFailure)
	err := front.SetWallpaper(State{DateCode: testDateString})
	if err.Error() != "exit status 5" {
		t.Fatal(err)
	}
}

func TestConfigurationE2e(t *testing.T) {
	testHome := setupTestHome(t)
	defer os.RemoveAll(testHome)
	cfg := "barewm"
	configFlag = &cfg
	if err := Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestJumpWithStateE2e(t *testing.T) {
	f, testHome := frontendForTestConfigured(t, nil)
	defer os.RemoveAll(testHome)
	makeStateFile(t)
	makeTestWallpapers(t, f.storage)
	writeWallpaperScript(setScriptSuccess)

	j := -1
	jump = &j
	nono := true
	nonotify = &nono

	if err := Execute(); err != nil {
		t.Fatal(err)
	}
}
