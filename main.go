package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/slspeek/apod-bg/apod"
)

var (
	info     = flag.Bool("info", false, "open the APOD-page on the current background")
	login    = flag.Bool("login", false, "do the procedure for a graphical login: download todays image and display it")
	logfile  = flag.String("log", os.ExpandEnv("${HOME}/.config/apod-bg/apod-bg.log"), "logfile specification")
	days     = flag.Int("fetch", 0, "days to go back downloading")
	jump     = flag.Int("jump", 0, "jump N backgrounds further, use negative numbers to jump backward")
	config   = flag.String("config", "", "initializes apod-bg for chosen window-manager")
	apodFlag = flag.Bool("apod", false, "opens the default browser on the Astronomy Picture of The Day")
	mode     = flag.Bool("mode", false, "mode background sizing options: fit or zoom")
)

func main() {
	flag.Parse()
	var a apod.APOD
	var logger *log.Logger
	{
		f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			fmt.Printf("Could not open logfile %q, because: %v\n", *logfile, err)
			return
		}
		defer f.Close()
		logger = log.New(f, "", log.LstdFlags|log.Lshortfile)
		a = apod.NewAPOD(logger)
	}
	var logf = func(format string, args ...interface{}) {
		fmt.Printf(format, args...)
		logger.Printf(format, args...)
	}

	if *config != "" {
		err := a.Configure(*config)
		if err != nil {
			logf("Could not write the configuration, because: %v\n", err)
			os.Exit(2)
		}
		logf("apod-bg was successfully configured\n")
	}

	if *apodFlag {
		err := a.OpenAPODToday()
		if err != nil {
			logf("Could not open the APOD page, because: %v\n", err)
			os.Exit(3)
		} else {
			logf("Opened the default browser on APOD\n")
		}
		return
	}

	if *info {
		err := a.OpenAPODOnBackground()
		if err != nil {
			logf("Could not open the APOD page on background now showing, because: %v\n", err)
			os.Exit(3)
		}
		logf("Opened the default browser on the APOD-page related to the current background image\n")
		return
	}

	if *login {
		today := a.Today()
		if downloaded, err := a.IsDownloaded(today); downloaded || err != nil {
			if err != nil {
				logf("Could not check whether today was downloaded, because: %v\n", err)
				os.Exit(4)
			}
			err := a.DisplayCurrent()
			if err != nil {
				logf("Today was already downloaded, but could not display the current wallpaper, because: %v\n", err)
				os.Exit(5)
			} else {
				logf("Displayed the current wallpaper, as today was already downloaded\n")
			}
			return
		}
		ok, err := a.Download(today)
		if err != nil {
			logger.Printf("An error occurred during todays (%s) image downloading: %v\n", today, err)
			os.Exit(6)
		}
		if !ok {
			logf("No new image today (%s) on APOD\n", today)
			err := a.DisplayCurrent()
			if err != nil {
				logf("Could not display the current wallpaper, because: %v\n", err)
				os.Exit(7)
			} else {
				logf("Displayed the current wallpaper, as today had no new image\n")
			}
			return
		}
		err = a.SetWallpaper(apod.State{DateCode: today, Options: "fit"})
		if err != nil {
			logf("Could not set the wallpaper to %s, because: %v\n", today, err)
			os.Exit(8)
		} else {
			logf("Wallpaper set to %s\n", today)
		}
		return
	}

	if *days > 0 {
		a.LoadRecentPast(*days)
	}

	if *jump != 0 {
		err := a.Jump(*jump)
		if err != nil {
			logf("Could not jump(%d): %v\n", *jump, err)
			os.Exit(9)
			return
		}
		logf("Jump was successfull\n")
		return
	}

	if *mode {
		m, err := a.ToggleViewMode()
		if err != nil {
			logf("Could not mode viewing options: %v\n", err)
			os.Exit(10)
		} else {
			logf("Inversed the viewing option to: %s\n", m)
		}
		return
	}
}
