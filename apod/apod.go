package apod

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/101loops/clock"
	"github.com/haklop/gnotifier"
	"github.com/skratchdot/open-golang/open"
)

const (
	apodBase           = "http://apod.nasa.gov/apod/"
	format             = "060102"
	imgprefix          = "apod-img-"
	stateFileBasename  = "now-showing"
	configFileBasename = "config.json"
	zoom               = "zoom"
	fit                = "fit"
)

func configDir() string {
	return os.ExpandEnv("${HOME}/.config/apod-bg")
}

func configFile() string {
	return filepath.Join(configDir(), configFileBasename)
}

func stateFile() string {
	return filepath.Join(configDir(), stateFileBasename)
}

func wallpaperSetScript() string {
	return filepath.Join(configDir(), "set-wallpaper.sh")
}

const setScriptBareWM = `#!/bin/bash
if test $WALLPAPER_OPTIONS = zoom; then
	feh --bg-fill $WALLPAPER
else
	feh --bg-max $WALLPAPER
fi
`
const setScriptLXDE = `#!/bin/bash
if test $WALLPAPER_OPTIONS = zoom; then
	pcmanfm --set-wallpaper=$WALLPAPER --wallpaper-mode=stretch
else
	pcmanfm --set-wallpaper=$WALLPAPER --wallpaper-mode=fit
fi
`

const setScriptGNOME = `#!/bin/bash
gsettings set org.gnome.desktop.background picture-uri file://$WALLPAPER
if test $WALLPAPER_OPTIONS = zoom; then
	gsettings set  org.gnome.desktop.background picture-options zoom
else
	gsettings set  org.gnome.desktop.background picture-options scaled
fi
gsettings set  org.gnome.desktop.background primary-color "000000"
gsettings set  org.gnome.desktop.background secondary-color "000000"
`
const configNotFound = "configuration file was not found. Please run apod-bg -config=<barewm|gnome|lxde> first, see man page for more information."

var imageExpr = regexp.MustCompile(`<a href="(.*\.(jpg|gif))">`)

var Notification = gnotifier.Notification

func Notify(mesg string) {
	notification := Notification("apod-bg", mesg)
	notification.GetConfig().Expiration = 3000
	notification.GetConfig().ApplicationName = "apod-bg"
	notification.Push()
}

// config sets where to find the wallpaper directory.
type config struct {
	WallpaperDir string
}

func (c *config) writeOut() error {
	bs, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile(), bs, 0644)
}

func (c *config) makeWallpaperDir() error {
	return os.MkdirAll(c.WallpaperDir, 0700)
}

func MakeConfigDir() error {
	err := os.MkdirAll(configDir(), 0700)
	if err != nil {
		return fmt.Errorf("Could not create configuration directory %q, because: %v\n", configDir(), err)
	}
	return nil
}

func writeWallpaperScript(script string) error {
	file := wallpaperSetScript()
	err := ioutil.WriteFile(file, []byte(script), 0755)
	return err
}

// APOD encapsulates communicating with apod.nasa.gov
type APOD struct {
	Clock  clock.Clock
	Config *config
	Client *http.Client
	Log    *log.Logger
}

// NewAPOD constructs a new APOD object
func NewAPOD(logger *log.Logger) APOD {
	cfg := new(config)
	a := APOD{Config: cfg,
		Clock:  clock.New(),
		Client: http.DefaultClient,
		Log:    logger}
	return a
}

// configure initializes the configuration according the config argument.
func (a *APOD) Configure(cfg string) error {
	a.Config = new(config)
	{
		err := MakeConfigDir()
		if err != nil {
			return err
		}
	}
	{
		a.Config.WallpaperDir = filepath.Join(configDir(), "wallpapers")
		err := a.Config.makeWallpaperDir()
		if err != nil {
			return err
		}
		err = a.Config.writeOut()
		if err != nil {
			return err
		}
	}
	script := ""
	switch cfg {
	case "barewm":
		script = setScriptBareWM
	case "lxde":
		script = setScriptLXDE
	case "gnome":
		script = setScriptGNOME
	default:
		return fmt.Errorf("Unknown configuration type: %s\n", cfg)
	}
	return writeWallpaperScript(script)
}

// Loadconfig loads a config from disk or, failing that, returns an error.
func (a *APOD) Loadconfig() error {
	cfgFile := configFile()
	cfgExists, err := exists(cfgFile)
	if err != nil {
		return err
	}
	if !cfgExists {
		return fmt.Errorf(configNotFound)
	}
	bs, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bs, a.Config)
	return err
}

// Today returns the date of today in a formatted string.
func (a *APOD) Today() string {
	t := a.Clock.Now()
	return t.Format(format)
}

// OpenAPOD opens the web page at apod.nasa.gov for a given day in the default browser.
func (a *APOD) OpenAPOD(isodate string) error {
	url := a.UrlForDate(isodate)
	return open.Start(url)
}

// OpenAPODToday opens the today's page at apod.nasa.gov
func (a *APOD) OpenAPODToday() error {
	return a.OpenAPOD(a.Today())
}

// OpenAPODOnBackground opens the apod.nasa.gov page for the wallpaper now showing, or failing that, throws an error.
func (a *APOD) OpenAPODOnBackground() error {
	s, err := a.State()
	if err != nil {
		return fmt.Errorf("Could not get hold on the picture that is currently shown, because: %v", err)
	}
	return a.OpenAPOD(s.DateCode)
}

// State defines the date of the apod-image being shown and display options
type State struct {
	DateCode string
	Options  string
}

// State returns the current State-struct read from disk, or today if there is no state file
func (a *APOD) State() (State, error) {
	present, err := exists(stateFile())
	if err != nil {
		return State{}, err
	}
	if !present {
		return State{DateCode: a.Today(), Options: fit}, nil
	}
	sfb, err := ioutil.ReadFile(stateFile())
	if err != nil {
		return State{}, err
	}
	var s State
	err = json.Unmarshal(sfb, &s)
	return s, err
}

// LoadRecentPast loads images from apod.nasa.gov to the wallpaper directory, for a set number of days back
func (a *APOD) LoadRecentPast(days int) {
	for _, isodate := range a.recentPast(days) {
		a.Download(isodate)
	}
}

// ContainsImage parses an APOD page for a linked image, returns success, and image URL if successful or an error
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
func (a *APOD) IsDownloaded(isodate string) (bool, error) {
	file := a.fileName(isodate)
	fileExists, err := exists(file)
	if err != nil {
		return false, err
	}
	return fileExists, nil
}

func (a *APOD) download(url string, isodate string) (string, error) {
	file := a.fileName(isodate)
	output, err := os.Create(file)
	if err != nil {
		return "", err
	}
	defer output.Close()
	resp, err := a.Client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	return file, err
}

// Download downloads the image from apod.nasa.gov for the given date.
func (a *APOD) Download(isodate string) (bool, error) {
	if downloaded, _ := a.IsDownloaded(isodate); downloaded {
		a.Log.Printf("Not downloading %s, it already exits\n", isodate)
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
	Notify(fmt.Sprintf("Downloading todays APOD-image"))
	file, err := a.download(imgURL, isodate)
	if err != nil {
		return true, err
	}
	a.Log.Printf("Successfully downloaded %s to %q\n", isodate, file)
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
	return 0, fmt.Errorf("%s was not found", isodate)
}

// UrlForDate returns the URL for the APOD page for the given ISO date.
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
	var idx int
	s, err := a.State()
	if err != nil {
		return err
	}
	idx, err = a.IndexOf(s.DateCode)
	if err != nil {
		return err
	}
	toGo := idx + n
	if toGo >= len(all) {
		return fmt.Errorf("End reached")
	}
	if toGo < 0 {
		return fmt.Errorf("Begin reached")
	}
	code := all[toGo][len(imgprefix):]
	st := State{DateCode: code, Options: fit}
	return a.SetWallpaper(st)
}

// SetWallpaper sets the wallpaper to the image from the wallpaper directory for the given date.
func (a *APOD) SetWallpaper(s State) error {
	wallpaper := a.fileName(s.DateCode)
	cmd := exec.Command(wallpaperSetScript())
	env := os.Environ()
	env = append(env, "WALLPAPER="+wallpaper)
	env = append(env, "WALLPAPER_OPTIONS="+s.Options)
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		return err
	}
	return a.store(s)
}

func (a *APOD) store(s State) error {
	f, err := os.Create(stateFile())
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	err = e.Encode(s)
	return err
}

// ToggleViewMode toggles the view mode fill/full.
func (a *APOD) ToggleViewMode() (string, error) {
	s, err := a.State()
	if err != nil {
		return "", err
	}
	if s.Options == fit {
		s.Options = zoom
	} else {
		s.Options = fit
	}
	return s.Options, a.SetWallpaper(s)
}

//DisplayCurrent reads the State file and sets the wallpaper accordingly.
func (a *APOD) DisplayCurrent() error {
	isodate, err := a.State()
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
