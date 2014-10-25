package apod

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/haklop/gnotifier"
	"github.com/skratchdot/open-golang/open"
)

var (
	info         = flag.Bool("info", false, "open the APOD-page on the current background")
	login        = flag.Bool("login", false, "do the procedure for APOD graphical login: download todays image and display it")
	logFileFlag  = flag.String("log", "", "logfile specification")
	days         = flag.Int("fetch", 0, "days to go back downloading")
	jump         = flag.Int("jump", 0, "jump N backgrounds further, use negative numbers to jump backward")
	configFlag   = flag.String("config", "", "initializes apod-bg for chosen window-manager")
	unconfigFlag = flag.Bool("unconfig", false, "removes the autostart entry for LXDE")
	apodFlag     = flag.Bool("apod", false, "opens the default browser on the Astronomy Picture of The Day")
	mode         = flag.Bool("mode", false, "mode background sizing options: fit or zoom")
	nonotify     = flag.Bool("nonotify", false, "do not send notifications to the desktop")
	noseed       = flag.Bool("noseed", false, "do not seed after configuring")
	dateFlag     = flag.String("date", "", "specify a date to be considered as now (for testing)")
	randomFlag   = flag.Bool("random", false, "pick a random archive picture")
)

const (
	stateFileBasename  = "now-showing"
	configFileBasename = "config.json"
	zoom               = "zoom"
	fit                = "fit"
)

const apodDesktop = `[Desktop Entry] 

Type=Application

Exec=apod-bg -login
`

const setScriptBareWM = `#!/bin/bash
if test $WALLPAPER_OPTIONS = zoom; then
	feh --bg-fill "$WALLPAPER"
else
	feh --bg-max "$WALLPAPER"
fi
`
const setScriptLXDE = `#!/bin/bash
if test $WALLPAPER_OPTIONS = zoom; then
	pcmanfm --set-wallpaper="$WALLPAPER" --wallpaper-mode=crop
else
	pcmanfm --set-wallpaper="$WALLPAPER" --wallpaper-mode=fit
fi
`

const setScriptGNOME = `#!/bin/bash
gsettings set org.gnome.desktop.background picture-uri "file://$WALLPAPER"
if test $WALLPAPER_OPTIONS = zoom; then
	gsettings set  org.gnome.desktop.background picture-options zoom
else
	gsettings set  org.gnome.desktop.background picture-options scaled
fi
gsettings set  org.gnome.desktop.background primary-color "000000"
gsettings set  org.gnome.desktop.background secondary-color "000000"
`
const configNotFound = "configuration file was not found. Please run apod-bg -config=barewm|gnome|lxde> first, see man page for more information."

func logFile() string {
	if *logFileFlag == "" {
		return os.ExpandEnv("${HOME}/.config/apod-bg/apod-bg.log")
	}
	return *logFileFlag
}

func configDir() string {
	return os.ExpandEnv("${HOME}/.config/apod-bg")
}

func autostartDir() string {
	return os.ExpandEnv("${HOME}/.config/autostart")
}

func autostartFile() string {
	return filepath.Join(autostartDir(), "apod-bg.desktop")
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

type logger interface {
	Printf(f string, i ...interface{})
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

func (c *config) fileName(isodate ADate) string {
	return filepath.Join(c.WallpaperDir, c.fileBaseName(isodate))
}

func (c *config) fileBaseName(isodate ADate) string {
	return fmt.Sprintf(imgPrefix+"%s", isodate.String())
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

type Notifier struct {
	Notification gnotifier.Builder
}

func (n *Notifier) Notify(mesg string) {
	notification := n.Notification("apod-bg", mesg)
	notification.GetConfig().Expiration = 3000
	notification.GetConfig().ApplicationName = "apod-bg"
	notification.Push()
}

type Frontend struct {
	Log    logger
	Config *config
	Notifier
	APOD    *APOD
	loader  *Loader
	storage *Storage
}

func NewFrontend(logger logger, notifier Notifier) *Frontend {
	APOD := NewAPOD()
	s := &Storage{}
	l := &Loader{APOD: APOD, logger: logger, Notifier: notifier}
	return &Frontend{
		Log:      logger,
		Notifier: notifier,
		APOD:     APOD,
		Config:   new(config),
		loader:   l,
		storage:  s}

}

// State defines the date of the apod-image being shown and display options
type State struct {
	DateCode ADate
	Options  string
}

// State returns the current State-struct read from disk, or APOD new State object set to today if there is no state file
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

func store(s State) error {
	fd, err := os.Create(stateFile())
	if err != nil {
		return err
	}
	defer fd.Close()
	e := json.NewEncoder(fd)
	err = e.Encode(s)
	return err
}

func (f *Frontend) writeAutostart() error {
	err := os.MkdirAll(autostartDir(), 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(autostartFile(), []byte(apodDesktop), 0644)
}

func (f *Frontend) removeAutostart() error {
	return os.Remove(autostartFile())
}

func (f *Frontend) Seed() error {
	if *noseed {
		return nil
	}
	date := f.Today()
	for i := 0; i < 7; i++ {
		loaded, err := f.loader.Download(date)
		date = *date.Back()
		if err != nil {
			continue
		}
		if loaded {
			break
		}
	}
	return f.RandomArchive()
}

// Today returns the date of today in APOD formatted string.
func (f *Frontend) Today() ADate {
	if *dateFlag != "" {
		return ADate(*dateFlag)
	}
	return NewADate(time.Now())
}

// configure initializes the configuration according the config argument.
func (f *Frontend) configure(cfg string) error {
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
		err := f.writeAutostart()
		if err != nil {
			return err
		}
	case "gnome":
		script = setScriptGNOME
	default:
		return fmt.Errorf("Unknown configuration type: %s\n", cfg)
	}
	if err := writeWallpaperScript(script); err != nil {
		return err
	}
	return f.Loadconfig()
}

// Configure initializes the configuration according the config argument and does seeding
// of images.
func (f *Frontend) Configure(cfg string) error {
	err := f.configure(cfg)
	if err != nil {
		return err
	}
	return f.Seed()
}

// Loadconfig loads APOD config from disk or, failing that, returns an error.
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
	f.loader.Config = f.Config
	return err
}

// OpenAPOD opens the web page at apod.nasa.gov for APOD given day in the default browser.
func (f *Frontend) OpenAPOD(isodate ADate) error {
	url := f.APOD.UrlForDate(ADate(isodate))
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

// Jump jumps to an image n places further or back if n is negative in the wallpaper directory.
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
	code := all[toGo]
	st := State{DateCode: code, Options: fit}
	return f.SetWallpaper(st)
}

// SetWallpaper sets the wallpaper to the image from the wallpaper directory for the given date.
func (f *Frontend) SetWallpaper(s State) error {
	wallpaper := f.Config.fileName(s.DateCode)
	cmd := exec.Command(wallpaperSetScript())
	env := os.Environ()
	env = append(env, "WALLPAPER="+wallpaper)
	env = append(env, "WALLPAPER_OPTIONS="+s.Options)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error running Wallpaper-Set-Script: %v. Output: %s", err, string(output))
	}
	return store(s)
}

// ToggleViewMode toggles the view mode fill/zoom. It returns the new state.
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

// DisplayCurrent reads the State file and sets the wallpaper accordingly.
func (f *Frontend) DisplayCurrent() error {
	isodate, err := f.State()
	if err != nil {
		return err
	}
	err = f.SetWallpaper(isodate)
	return err
}

func initLogging() (*log.Logger, *os.File, error) {
	var logger *log.Logger
	err := MakeConfigDir()
	if err != nil {
		return nil, nil, fmt.Errorf("Could not create config dir")
	}
	f, err := os.OpenFile(logFile(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)

	if err != nil {
		return nil, f, fmt.Errorf("Could not open logfile %q, because: %v\n", logFile(), err)
	}
	mw := io.MultiWriter(os.Stdout, f)

	logger = log.New(mw, "", log.LstdFlags)
	return logger, f, nil
}

// RunAtLogin should be configured to run when the user starts her
// windowmanager. It checks for a new APOD image. If there is a new
// image is sets this as background, otherwise it display a random
// archive image.
func (f *Frontend) RunAtLogin() error {
	today := f.Today()
	if downloaded, err := f.Config.IsDownloaded(today); downloaded || err != nil {
		if err != nil {
			return fmt.Errorf("Could not check whether today was downloaded, because: %v\n", err)
		}
		err := f.DisplayCurrent()
		if err != nil {
			return fmt.Errorf("Today was already downloaded, but could not display the current wallpaper, because: %v\n", err)
		} else {
			f.Log.Printf("Displayed the current wallpaper, as today was already downloaded\n")
		}
		return nil
	}
	ok, err := f.loader.Download(today)
	if err != nil {
		f.Log.Printf("An error occurred during todays (%s) image downloading: %v\n", today, err)
		// The show must go on
	}
	if !ok {
		f.Log.Printf("No new image today (%s) on APOD\n", today)

		f.Notify(fmt.Sprintf("No new image today :-("))
		err := f.RandomArchive()
		if err != nil {
			return fmt.Errorf("Could not display a random archive, because: %v\n", err)
		} else {
			f.Log.Printf("Displayed a random archive wallpaper, as today had no new image\n")
		}
		return nil
	}
	err = f.SetWallpaper(State{DateCode: today, Options: "fit"})
	if err != nil {
		return fmt.Errorf("Could not set the wallpaper to %s, because: %v\n", today, err)
	} else {
		mesg := fmt.Sprintf("Wallpaper set to %s\n", today)
		f.Notify(mesg)
		f.Log.Printf(mesg)
	}
	return nil
}

// RandomArchive picks a random image from the already downloaded
// images
func (f *Frontend) RandomArchive() error {
	rand.Seed(time.Now().Unix())
	bs, err := f.storage.DownloadedWallpapers()
	if err != nil {
		return err
	}
	n := len(bs)
	if n == 0 {
		return fmt.Errorf("No backgrounds downloaded yet")
	}
	if n > 1 {
		// Don't want yesterdays wallpaper
		n -= 1
	}
	pick := bs[rand.Intn(n)]
	s, err := f.State()
	if err != nil {
		return err
	}
	s.DateCode = pick
	return f.SetWallpaper(s)
}

// Execute is the entry point for the apod-bg command
func Execute() error {
	logger, f, err := initLogging()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()
	var front *Frontend
	logger.Printf("apod-bg starts")
	if *nonotify {
		front = NewFrontend(logger, Notifier{gnotifier.NullNotification})
	} else {
		front = NewFrontend(logger, Notifier{gnotifier.Notification})
	}
	if *configFlag != "" {
		err := front.Configure(*configFlag)
		if err != nil {
			err = fmt.Errorf("Could not properly configure the apod-bg, because: %v\n", err)
			logger.Printf("%v\n", err)
			return err
		}
		logger.Printf("apod-bg was successfully configured\n")
		return nil
	}
	err = front.Loadconfig()
	if err != nil {
		err = fmt.Errorf("Could not load the configuration, because: %v\n", err)
		logger.Printf("%v\n", err)
		return err
	}

	if *apodFlag {
		err := front.OpenAPODToday()
		if err != nil {
			err = fmt.Errorf("Could not open the APOD page, because: %v\n", err)
			logger.Printf("%v\n", err)
			return err
		} else {
			mesg := "Opened the default browser on APOD\n"
			front.Notify(mesg)
			logger.Printf(mesg)
		}
		return nil
	}

	if *unconfigFlag {
		err := front.removeAutostart()
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}
	}

	if *randomFlag {
		err := front.RandomArchive()
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}
	}

	if *info {
		err := front.OpenAPODOnBackground()
		if err != nil {
			err = fmt.Errorf("Could not open the APOD page on background now showing, because: %v\n", err)
			logger.Printf("%v\n", err)
			return err
		}
		logger.Printf("Opened the default browser on the APOD-page related to the current background image\n")
		front.Notify("Browser opened on NASA apod-page belonging to this background")
		return nil
	}

	if *login {
		err := front.RunAtLogin()
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}
	}

	if *days > 0 {
		front.loader.LoadPeriod(front.Today(), *days)
	}

	if *jump != 0 {
		err := front.Jump(*jump)
		if err != nil {
			front.Notify(err.Error())
			err = fmt.Errorf("Could not jump(%d): %v\n", *jump, err)
			logger.Printf("%v\n", err)
			return err
		}
		logger.Printf("Jump was successfull\n")
		return nil
	}

	if *mode {
		m, err := front.ToggleViewMode()
		if err != nil {
			err = fmt.Errorf("Could not toggle viewing options: %v\n", err)
			logger.Printf("%v\n", err)
			return err

		} else {
			logger.Printf("Inversed the viewing option to: %s\n", m)
		}
		return nil
	}
	return nil
}
