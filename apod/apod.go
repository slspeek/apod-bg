package apod

import (
	"github.com/101loops/clock"
)

const format = "060102"

type Config struct {
	Fehbg        string
	WallpaperDir string
}

func LoadConfig() (Config, error) {
	return Config{}, nil
}

type APOD struct {
	Clock  clock.Clock
	Config Config
}

// Now returns a string with the date in ISO format.
func (a *APOD) Now() string {
	t := a.Clock.Now()
	return t.Format(format)
}

// These are placeholder functions that are required to compile.
func (a *APOD) Today() string {
	return a.Now()
}

func (a *APOD) NowShowing() string {
	return "The background image showing now"
}

func (a *APOD) RecentHistory(days int) string {
	return "The last N days"
}

func (a *APOD) ContainsImage(url string) bool {
	return false
}

func (a *APOD) IsDownloaded(isodate string) bool {
	return false
}

func (a *APOD) Download(url string, isodate string) error {
	return nil
}

func (a *APOD) IndexOf(isodate string) (int, error) {
	return 0, nil
}

func (a *APOD) SetViewingMode(fill bool) {}

func (a *APOD) UrlForDate(isodate string) string {
	return "Some APOD URL"
}

func (a *APOD) DownloadedWallpapers() string {
	return "Downloaded wallpapers"
}

func (a *APOD) SetWallpaper(path string) {}
