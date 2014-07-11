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
	"sort"
)

const apodBase = "http://apod.nasa.gov/apod/"
const format = "060102"
const setWallpaperScript = `#!/bin/bash
feh --bg-scale $WALLPAPER
`

var imageExpr = regexp.MustCompile(`<a href="(.*\.(jpg|gif))">`)

type Config struct {
	StateFile    string
	WallpaperDir string
}

func LoadConfig() (Config, error) {
	configDir := os.ExpandEnv("${HOME}/.config/apod-bg")
	err := os.MkdirAll(configDir, 0700)
	if err != nil {
		return Config{}, err
	}
	setScript := filepath.Join(configDir, "set-wallpaper.sh")
	ok, err := exists(setScript)
	if err != nil {
		return Config{}, err
	}
	if !ok {
		s, err := os.Create(setScript)
		if err != nil {
			return Config{}, err
		}
		_, err = s.WriteString(setWallpaperScript)
		if err != nil {
			return Config{}, err
		}
	}
	nowShowing := filepath.Join(configDir, "now-showing")
	wallpaperDir := filepath.Join(configDir, "wallpapers")
	err = os.MkdirAll(wallpaperDir, 0700)
	if err != nil {
		return Config{}, err
	}
	return Config{StateFile: nowShowing, WallpaperDir: wallpaperDir}, nil
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
	return string(bs), nil
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
	target := a.fileBaseName(isodate)
	all, err := a.DownloadedWallpapers()
	if err != nil {
		return 0, err
	}
	for i, elem := range all {
		if elem == target {
			return i, nil
		}
	}
	return 0, fmt.Errorf("Not found")
}

func (a *APOD) SetViewingMode(fill bool) {}

// UrlForDate returns the URL for the APOD page for the given isodate.
func (a *APOD) UrlForDate(isodate string) string {
	return fmt.Sprintf("http://apod.nasa.gov/apod/ap%s.html", isodate)
}

func (a *APOD) fileName(isodate string) string {
	return filepath.Join(a.Config.WallpaperDir, a.fileBaseName(isodate))
}

func (a *APOD) fileBaseName(isodate string) string {
	return fmt.Sprintf("apod-img-%s", isodate)
}

func (a *APOD) DownloadedWallpapers() ([]string, error) {
	dir, err := os.Open(a.Config.WallpaperDir)
	if err != nil {
		return nil, err
	}
	files, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func (a *APOD) Jump(n int) error {
	return nil
}

func (a *APOD) SetWallpaper(isodate string) error {
	return nil
}

func (a *APOD) ToggleViewMode() {
}

func (a *APOD) VisitAPODSite() {
}

func (a *APOD) DisplayCurrent() error {
	isodate, err := a.NowShowing()
	if err != nil {
		return err
	}
	err = a.SetWallpaper(isodate)
	return err
}
