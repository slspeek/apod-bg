package apod

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/101loops/clock"
	"github.com/skratchdot/open-golang/open"
)

const (
	apodBase  = "http://apod.nasa.gov/apod/"
	format    = "060102"
	imgprefix = "apod-img-"
)

const setWallpaperScriptBareWM = `#!/bin/bash
feh --bg-max $WALLPAPER
`

var imageExpr = regexp.MustCompile(`<a href="(.*\.(jpg|gif))">`)

// Config sets where to find the image now showing, the wallpaper directory and the wallpaper to set.
type Config struct {
	StateFile    string
	WallpaperDir string
	SetWallpaper string
}

// LoadConfig loads the above Config or, failing that, throw an error.
func LoadConfig() (Config, error) {
	configDir := os.ExpandEnv("${HOME}/.config/apod-bg")
	err := os.MkdirAll(configDir, 0700)
	if err != nil {
		return Config{}, err
	}
	scriptFile := filepath.Join(configDir, "set-wallpaper.sh")
	ok, err := exists(scriptFile)
	if err != nil {
		return Config{}, err
	}
	if !ok {
		s, err := os.Create(scriptFile)
		if err != nil {
			return Config{}, err
		}
		_, err = s.WriteString(setWallpaperScriptBareWM)
		if err != nil {
			return Config{}, err
		}
		err = s.Close()
		if err != nil {
			return Config{}, err
		}
		err = os.Chmod(scriptFile, 0755)
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
	return Config{StateFile: nowShowing, WallpaperDir: wallpaperDir, SetWallpaper: scriptFile}, nil
}

type APOD struct {
	Clock  clock.Clock
	Config Config
	Client *http.Client
}

// Today returns the date of today in a formatted string.
func (a *APOD) Today() string {
	t := a.Clock.Now()
	return t.Format(format)
}

// OpenAPOD opens the webpage at apod.nasa.gov for a given day in the default browser.
func (a *APOD) OpenAPOD(isodate string) error {
	url := a.UrlForDate(isodate)
	return open.Start(url)
}

// OpenAPODToday opens the today's page at apod.nasa.gov .
func (a *APOD) OpenAPODToday() error {
	return a.OpenAPOD(a.Today())
}

// OpenAPODOnBackground opens the apod.nasa.gov page for the wallpaper now showing, or failing that, throws an error.
func (a *APOD) OpenAPODOnBackground() error {
	isodate, err := a.NowShowing()
	if err != nil {
		return fmt.Errorf("Could not get hold on the picture that is currently shown, because: %v", err)
	}
	return a.OpenAPOD(isodate)
}

// NowShowing returns isodate string YYMMDD.
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

// LoadRecentPast loads images from apod.nasa.gov to the wallpaper dir, for a set number of days back, or throws error.
func (a *APOD) LoadRecentPast(days int) {
	for _, isodate := range a.recentPast(days) {
		a.Download(isodate)
	}
}

// ContainsImage parses an APOD page for a linked image, returns false/true, and image URL if successful.
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

// IsDownloaded checks whether an image file is downloaded for a given date.
func (a *APOD) IsDownloaded(isodate string) bool {
	file := a.fileName(isodate)
	fileExists, err := exists(file)
	if err != nil {
		return true
	}
	return fileExists
}

func (a *APOD) download(url string, isodate string) error {
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

// Download downloads the image from apod.nasa.gov for the given date.
func (a *APOD) Download(isodate string) (bool, error) {
	if downloaded := a.IsDownloaded(isodate); downloaded {
		return true, nil
	}
	pageURL := a.UrlForDate(isodate)
	success, imgURL, err := a.ContainsImage(pageURL)
	if err != nil {
		return false, err
	}
	if !success {
		return false, nil
	}
	err = a.download(imgURL, isodate)
	if err != nil {
		return true, err
	}
	return true, nil
}

// IndexOf returns the index of an image in the wallpaper directory.
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

// Jump jumps to an image next or previous in the wallpaper directory.
func (a *APOD) Jump(n int) error {
	all, err := a.DownloadedWallpapers()
	if err != nil {
		return err
	}
	now, err := a.NowShowing()
	if err != nil {
		return err
	}
	idx, err := a.IndexOf(now)
	if err != nil {
		return err
	}
	toGo := idx + n
	if toGo >= len(all) || toGo < 0 {
		return fmt.Errorf("Out of bounds")
	}
	newImageBaseName := all[toGo]
	err = a.SetWallpaper(newImageBaseName[9:])
	return err
}

// SetWallpaper sets the wallpaper to the image from the wallpaper directory for the given date.
func (a *APOD) SetWallpaper(isodate string) error {
	wallpaper := a.fileName(isodate)
	cmd := exec.Command(a.Config.SetWallpaper)
	env := os.Environ()
	env = append(env, "WALLPAPER="+wallpaper)
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		return err
	}
	return a.store(isodate)
}

func (a *APOD) store(isodate string) error {
	s, err := os.Create(a.Config.StateFile)
	if err != nil {
		return err
	}
	_, err = s.WriteString(isodate)
	return err
}

// ToggleViewMode toggles the view mode fill/full.
func (a *APOD) ToggleViewMode() {
}

//DisplayCurrent reads the state file and sets the wallpaper accordingly.
func (a *APOD) DisplayCurrent() error {
	isodate, err := a.NowShowing()
	if err != nil {
		return err
	}
	err = a.SetWallpaper(isodate)
	return err
}

func (a *APOD) recentPast(days int) []string {
	today := a.Clock.Now()
	var dates []string
	for i := 1; i < days+1; i++ {
		dates = append(dates, today.AddDate(0, 0, -i).Format(format))
	}
	return dates
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

func (a *APOD) fileName(isodate string) string {
	return filepath.Join(a.Config.WallpaperDir, a.fileBaseName(isodate))
}

func (a *APOD) fileBaseName(isodate string) string {
	return fmt.Sprintf(imgprefix+"%s", isodate)
}
