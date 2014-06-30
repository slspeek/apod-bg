package apod

import (
	"github.com/101loops/clock"
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	iso := "140121"
	m := clock.NewMock()
	m.Set(t0)

	apod := APOD{m}
	if apod.Now() != iso {
		t.Errorf("Expected %v, got %v", iso, apod.Now())
	}
}

func TestToday(t *testing.T) {
	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	iso := "140121"
	m := clock.NewMock()
	m.Set(t0)

	apod := APOD{m}
	if apod.Today() != iso {
		t.Errorf("Expected %v, got %v", iso, apod.Today())
	}
}
