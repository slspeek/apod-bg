package apod

import (
	"github.com/101loops/clock"
)

const format = "20060102"

type APOD struct {
	Clock clock.Clock
}

// Now retutns a string with the date in ISO format.
func (a *APOD) Now() string {
  t := a.Clock.Now()
  return t.Format(format) 
}
