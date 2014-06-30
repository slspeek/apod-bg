package apod

import (
	"github.com/101loops/clock"
)

const format = "060102"

type APOD struct {
	Clock clock.Clock
}

// Now returns a string with the date in ISO format.
func (a *APOD) Now() string {
	t := a.Clock.Now()
	return t.Format(format)
}

// These are placeholder functions that are required to compile.
func Today() string {
  return "Today"
}

func NowShowing() string {
  return "The background image showing now"
}

func RecentHistory(days int) string {
  return "The last N days"
}

func ContainsImage(url string) bool {
  return false
}

func IsDownloaded(isodate string) bool {
  return false
}

func Download(url string, isodate string) error {
  return nil
}

func IndexOf(isodate string) (int, error) {
  return 0, nil
}

func SetViewingMode(fill bool) {}

func LoadConfig() error {
  return nil
}

func WallpaperDir() string {
  return "Wallpaper directory"
}

func UrlForDate(isodate string) string {
  return "Some APOD URL"
}

func DownloadedWallpapers() string {
  return "Downloaded wallpapers"
}

func SetWallpaper(path string) {}
