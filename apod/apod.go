package apod

import (
	"fmt"
	"github.com/101loops/clock"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

const format = "060102"

var isodateExpr = regexp.MustCompile(`(\d{6})`)

type Config struct {
	StateFile    string
	WallpaperDir string
}

func LoadConfig() (Config, error) {
	fehbg := os.ExpandEnv("${HOME}/.fehbg")
	wallpaperDir := os.ExpandEnv("${HOME}/.config/apod-bg/wallpapers")
	err := os.MkdirAll(wallpaperDir, 0700)
	if err != nil {
		return Config{}, err
	}
	return Config{StateFile: fehbg, WallpaperDir: wallpaperDir}, nil
}

type APOD struct {
	Clock  clock.Clock
	Config Config
	Client *http.Client
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

func (a *APOD) NowShowing() (string, error) {
	sf, err := os.Open(a.Config.StateFile)
	if err != nil {
		return "", err
	}
	bs, err := ioutil.ReadAll(sf)
	if err != nil {
		return "", err
	}
	if m := isodateExpr.FindStringSubmatch(string(bs)); m != nil {
		return m[1], nil
	} else {
		return "", fmt.Errorf("Nothing found")
	}
}

func (a *APOD) RecentHistory(days int) []string {
	return []string{""}
}

func (a *APOD) loadPage(url string) (string, error) {
	return "", nil
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

// UrlForDate returns the URL for the APOD page for the given isodate.
func (a *APOD) UrlForDate(isodate string) string {
	return fmt.Sprintf("http://apod.nasa.gov/apod/ap%s.html", isodate)
}

func (a *APOD) DownloadedWallpapers() []string {
	return []string{""}
}

func (a *APOD) SetWallpaper(path string) {}
