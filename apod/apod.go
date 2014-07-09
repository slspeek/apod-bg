package apod

import (
	"fmt"
	"github.com/101loops/clock"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const apodBase = "http://apod.nasa.gov/apod/"
const format = "060102"

var isodateExpr = regexp.MustCompile(`(\d{6})`)
var imageExpr = regexp.MustCompile(`<a href="(.*\.(jpg|gif))">`)

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

// Today returns a string with the date in ISO format.
func (a *APOD) Today() string {
	t := a.Clock.Now()
	return t.Format(format)
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
	resp, err := a.Client.Get(url)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (a *APOD) ContainsImage(url string) (bool, string, error) {
	content, err := a.loadPage(url)
	if err != nil {
		return false, "", err
	}
	m := imageExpr.FindStringSubmatch(content)
	if m != nil && m[1] != "" {
		return true, apodBase + m[1], nil
	}
	return false, "", nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (a *APOD) IsDownloaded(isodate string) bool {
	file := a.fileName(isodate)
	fileExists, err := exists(file)
	if err != nil {
		return true
	}
	return fileExists
}

func (a *APOD) Download(url string, isodate string) error {
	file := a.fileName(isodate)
	output, err := os.Create(file)
	if err != nil {
		return err
	}
	defer output.Close()
	resp, err := a.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	return err
}

func (a *APOD) IndexOf(isodate string) (int, error) {
	return 0, nil
}

func (a *APOD) SetViewingMode(fill bool) {}

// UrlForDate returns the URL for the APOD page for the given isodate.
func (a *APOD) UrlForDate(isodate string) string {
	return fmt.Sprintf("http://apod.nasa.gov/apod/ap%s.html", isodate)
}

func (a *APOD) fileName(isodate string) string {
	return filepath.Join(a.Config.WallpaperDir, fmt.Sprintf("apod-img-%s", isodate))
}

func (a *APOD) DownloadedWallpapers() []string {
	return []string{""}
}

func (a *APOD) SetWallpaper(path string) {}
