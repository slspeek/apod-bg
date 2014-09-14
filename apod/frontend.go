package apod

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/101loops/clock"
	"github.com/haklop/gnotifier"
	"github.com/skratchdot/open-golang/open"
)

var (
	info       = flag.Bool("info", false, "open the APOD-page on the current background")
	login      = flag.Bool("login", false, "do the procedure for a graphical login: download todays image and display it")
	logfile    = flag.String("log", os.ExpandEnv("${HOME}/.config/apod-bg/apod-bg.log"), "logfile specification")
	days       = flag.Int("fetch", 0, "days to go back downloading")
	jump       = flag.Int("jump", 0, "jump N backgrounds further, use negative numbers to jump backward")
	configFlag = flag.String("config", "", "initializes apod-bg for chosen window-manager")
	apodFlag   = flag.Bool("apod", false, "opens the default browser on the Astronomy Picture of The Day")
	mode       = flag.Bool("mode", false, "mode background sizing options: fit or zoom")
	nonotify   = flag.Bool("nonotify", false, "do not send notifications to the desktop")
)

const (
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

type Frontend struct {
	Clock    clock.Clock
	Log      *log.Logger
	Config   *config
	Notifier gnotifier.Builder
	a        *APOD
	loader   *Loader
	storage  *Storage
}

func NewFrontend(logger *log.Logger) *Frontend {
	a := NewAPOD()
	s := &Storage{}
	l := &Loader{a: a, storage: s}
	return &Frontend{Clock: clock.New(),
		Log:      logger,
		Notifier: gnotifier.Notification,
		a:        a,
		loader:   l,
		storage:  s}

}

// State defines the date of the apod-image being shown and display options
type State struct {
	DateCode string
	Options  string
}

// State returns the current State-struct read from disk, or today if there is no state file
func (f *Frontend) State() (State, error) {
	present, err := exists(stateFile())
	if err != nil {
		return State{}, err
	}
	if !present {
		return State{DateCode: f.Today(), Options: fit}, nil
	}
	sfb, err := ioutil.ReadFile(stateFile())
	if err != nil {
		return State{}, err
	}
	var s State
	err = json.Unmarshal(sfb, &s)
	return s, err
}

func (f *Frontend) store(s State) error {
	fd, err := os.Create(stateFile())
	if err != nil {
		return err
	}
	defer fd.Close()
	e := json.NewEncoder(fd)
	err = e.Encode(s)
	return err
}

func (f *Frontend) Notify(mesg string) {
	notification := f.Notifier("apod-bg", mesg)
	notification.GetConfig().Expiration = 3000
	notification.GetConfig().ApplicationName = "apod-bg"
	notification.Push()
}

// Today returns the date of today in a formatted string.
func (f *Frontend) Now() time.Time {
	return f.Clock.Now()
}

// Today returns the date of today in a formatted string.
func (f *Frontend) Today() string {
	t := f.Now()
	return t.Format(format)
}

// configure initializes the configuration according the config argument.
func (f *Frontend) Configure(cfg string) error {
	f.Config = new(config)
	{
		err := MakeConfigDir()
		if err != nil {
			return err
		}
	}
	{
		f.Config.WallpaperDir = filepath.Join(configDir(), "wallpapers")
		err := f.Config.makeWallpaperDir()
		if err != nil {
			return err
		}
		err = f.Config.writeOut()
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
func (f *Frontend) Loadconfig() error {
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
	err = json.Unmarshal(bs, f.Config)
	if err != nil {
		return err
	}
	f.storage.Config = f.Config
	return err
}

// OpenAPOD opens the web page at apod.nasa.gov for a given day in the default browser.
func (f *Frontend) OpenAPOD(isodate string) error {
	url := f.a.UrlForDate(isodate)
	return open.Start(url)
}

// OpenAPODToday opens the today's page at apod.nasa.gov
func (f *Frontend) OpenAPODToday() error {
	return f.OpenAPOD(f.Today())
}

// OpenAPODOnBackground opens the apod.nasa.gov page for the wallpaper now showing, or failing that, throws an error.
func (f *Frontend) OpenAPODOnBackground() error {
	s, err := f.State()
	if err != nil {
		return fmt.Errorf("Could not get hold on the picture that is currently shown, because: %v", err)
	}
	return f.OpenAPOD(s.DateCode)
}

// Jump jumps to an image next or previous in the wallpaper directory.
func (f *Frontend) Jump(n int) error {
	all, err := f.storage.DownloadedWallpapers()
	if err != nil {
		return err
	}
	var idx int
	s, err := f.State()
	if err != nil {
		return err
	}
	idx, err = f.storage.IndexOf(s.DateCode)
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
	return f.SetWallpaper(st)
}

// SetWallpaper sets the wallpaper to the image from the wallpaper directory for the given date.
func (f *Frontend) SetWallpaper(s State) error {
	wallpaper := f.storage.fileName(s.DateCode)
	cmd := exec.Command(wallpaperSetScript())
	env := os.Environ()
	env = append(env, "WALLPAPER="+wallpaper)
	env = append(env, "WALLPAPER_OPTIONS="+s.Options)
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		return err
	}
	return f.store(s)
}

// ToggleViewMode toggles the view mode fill/full.
func (f *Frontend) ToggleViewMode() (string, error) {
	s, err := f.State()
	if err != nil {
		return "", err
	}
	if s.Options == fit {
		s.Options = zoom
	} else {
		s.Options = fit
	}
	return s.Options, f.SetWallpaper(s)
}

//DisplayCurrent reads the State file and sets the wallpaper accordingly.
func (f *Frontend) DisplayCurrent() error {
	isodate, err := f.State()
	if err != nil {
		return err
	}
	err = f.SetWallpaper(isodate)
	return err
}

// Execute is the entry point for the apod-bg command
func Execute() {
	var front *Frontend
	var logger *log.Logger
	{
		err := MakeConfigDir()
		if err != nil {
			fmt.Printf("Could not create config dir")
			os.Exit(1)
		}
		f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			fmt.Printf("Could not open logfile %q, because: %v\n", *logfile, err)
			os.Exit(2)
		}
		defer f.Close()
		mw := io.MultiWriter(os.Stdout, f)

		logger = log.New(mw, "", log.LstdFlags)
		front = NewFrontend(logger)
	}
	if *nonotify {
		front.Notifier = gnotifier.NullNotification
	}
	if *configFlag != "" {
		err := front.Configure(*configFlag)
		if err != nil {
			logger.Fatalf("Could not write the configuration, because: %v\n", err)
		}
		logger.Printf("apod-bg was successfully configured\n")
		return
	}
	err := front.Loadconfig()
	if err != nil {
		logger.Fatalf("Could not load the configuration, because: %v\n", err)
	}

	if *apodFlag {
		err := front.OpenAPODToday()
		if err != nil {
			logger.Fatalf("Could not open the APOD page, because: %v\n", err)
		} else {
			mesg := "Opened the default browser on APOD\n"
			front.Notify(mesg)
			logger.Printf(mesg)
		}
		return
	}

	if *info {
		err := front.OpenAPODOnBackground()
		if err != nil {
			logger.Fatalf("Could not open the APOD page on background now showing, because: %v\n", err)
		}
		logger.Printf("Opened the default browser on the APOD-page related to the current background image\n")
		front.Notify("Browser opened on NASA apod-page belonging to this background")
		return
	}

	if *login {
		today := front.Today()
		if downloaded, err := front.storage.IsDownloaded(today); downloaded || err != nil {
			if err != nil {
				logger.Fatalf("Could not check whether today was downloaded, because: %v\n", err)
			}
			err := front.DisplayCurrent()
			if err != nil {
				logger.Fatalf("Today was already downloaded, but could not display the current wallpaper, because: %v\n", err)
			} else {
				logger.Printf("Displayed the current wallpaper, as today was already downloaded\n")
			}
			return
		}
		ok, err := front.loader.Download(today)
		if err != nil {
			logger.Printf("An error occurred during todays (%s) image downloading: %v\n", today, err)
		}
		if !ok {
			logger.Printf("No new image today (%s) on APOD\n", today)

			front.Notify(fmt.Sprintf("No new image today :-("))
			err := front.DisplayCurrent()
			if err != nil {
				logger.Fatalf("Could not display the current wallpaper, because: %v\n", err)
			} else {
				logger.Printf("Displayed the current wallpaper, as today had no new image\n")
			}
			return
		}
		err = front.SetWallpaper(State{DateCode: today, Options: "fit"})
		if err != nil {
			logger.Fatalf("Could not set the wallpaper to %s, because: %v\n", today, err)
		} else {
			mesg := fmt.Sprintf("Wallpaper set to %s\n", today)
			front.Notify(mesg)
			logger.Printf(mesg)
		}
		return
	}

	if *days > 0 {
		front.loader.LoadRecentPast(front.Now(), *days)
	}

	if *jump != 0 {
		err := front.Jump(*jump)
		if err != nil {
			front.Notify(err.Error())
			logger.Fatalf("Could not jump(%d): %v\n", *jump, err)
		}
		logger.Printf("Jump was successfull\n")
		return
	}

	if *mode {
		m, err := front.ToggleViewMode()
		if err != nil {
			logger.Fatalf("Could not mode viewing options: %v\n", err)
		} else {
			logger.Printf("Inversed the viewing option to: %s\n", m)
		}
		return
	}
}
