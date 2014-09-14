package apod

import (
	"github.com/101loops/clock"
	"github.com/haklop/gnotifier"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func frontendForTest(t *testing.T, tripper http.RoundTripper) (*Frontend, string) {
	recorder := gnotifier.NewTestRecorder()
	f := new(Frontend)
	f.Notifier = recorder.Notification

	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	m := clock.NewMock()
	m.Set(t0)
	f.Clock = m

	f.Log = log.New(os.Stdout, "", log.LstdFlags)

	a := new(APOD)
	a.Client = &http.Client{Transport: tripper}
	f.a = a

	s := new(Storage)
	l := &Loader{a: a, storage: s}
	f.loader = l
	f.storage = s

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

func TestJumpWithoutState(t *testing.T) {
	f, testHome := frontendForTestConfigured(t, imageRoundTrip{})
	defer os.RemoveAll(testHome)
	prepareTest(t, f.storage)
	writeWallpaperScript(setScriptSuccess)
	err := f.Jump(-1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestState(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, nil)
	defer os.RemoveAll(testHome)
	err := ioutil.WriteFile(stateFile(), []byte(`{"DateCode":"140121","Options":"fit"}`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	rv, err := a.State()
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
	a, testHome := frontendForTestConfigured(t, nil)
	defer os.RemoveAll(testHome)

	err := a.Loadconfig()
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

func TestConfigurationE2e(t *testing.T) {
	testHome := setupTestHome(t)
	defer os.RemoveAll(testHome)
	cfg := "barewm"
	configFlag = &cfg
	Execute()
}

func TestSetWallpaperSuccess(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, testRoundTrip{})
	defer os.RemoveAll(testHome)
	writeWallpaperScript(setScriptSuccess)
	err := a.SetWallpaper(State{DateCode: testDateString})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetWallpaperFailure(t *testing.T) {
	a, testHome := frontendForTestConfigured(t, testRoundTrip{})
	defer os.RemoveAll(testHome)
	writeWallpaperScript(setScriptFailure)
	err := a.SetWallpaper(State{DateCode: testDateString})
	if err.Error() != "exit status 5" {
		t.Fatal(err)
	}
}
