package apod

import (
	"github.com/101loops/clock"
	"os"
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	iso := "140121"
	m := clock.NewMock()
	m.Set(t0)

	apod := APOD{Clock: m}
	if apod.Now() != iso {
		t.Errorf("Expected %v, got %v", iso, apod.Now())
	}
}

func TestToday(t *testing.T) {
	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	iso := "140121"
	m := clock.NewMock()
	m.Set(t0)

	apod := APOD{Clock: m}
	if apod.Today() != iso {
		t.Errorf("Expected %v, got %v", iso, apod.Today())
	}
}

func TestUrlForDate(t *testing.T) {
	apod := APOD{}
	url := apod.UrlForDate("140121")
	if url != "http://apod.nasa.gov/apod/ap140121.html" {
		t.Errorf("Expected: http://apod.nasa.gov/apod/ap140121.html, got %s", url)
	}
}

func TestNowShowing(t *testing.T) {
	f, err := os.Create("foo")
	if err != nil {
		t.Fatal("could not create file foo")
	}
	defer os.Remove("foo")
	_, err = f.WriteString(`feh  --bg-max '140121-apod.jpg'`)
	if err != nil {
		t.Fatal("could write to file  foo")
	}
	err = f.Close()
	if err != nil {
		t.Fatal("could close file foo")
	}
	cfg := Config{StateFile: "foo"}

	apod := APOD{Config: cfg}

	rv, err := apod.NowShowing()
	if err != nil {
		t.Fatalf("Error during call to NowShowing")
	}
	if rv != "140121" {
		t.Errorf("Expected 140121, got %v", rv)
	}
}
