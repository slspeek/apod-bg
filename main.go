package main

import (
	"flag"
	"fmt"
	"io"
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
		err := apod.MakeConfigDir()
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
		a = apod.NewAPOD(logger)
	}
	if *config != "" {
		err := a.Configure(*config)
		if err != nil {
			logger.Fatalf("Could not write the configuration, because: %v\n", err)
		}
		logger.Printf("apod-bg was successfully configured\n")
		return
	}
	err := a.Loadconfig()
	if err != nil {
		logger.Fatalf("Could not load the configuration, because: %v\n", err)
	}

	if *apodFlag {
		err := a.OpenAPODToday()
		if err != nil {
			logger.Fatalf("Could not open the APOD page, because: %v\n", err)
		} else {
			mesg := "Opened the default browser on APOD\n"
			apod.Notify(mesg)
			logger.Printf(mesg)
		}
		return
	}

	if *info {
		err := a.OpenAPODOnBackground()
		if err != nil {
			logger.Fatalf("Could not open the APOD page on background now showing, because: %v\n", err)
		}
		logger.Printf("Opened the default browser on the APOD-page related to the current background image\n")
		apod.Notify("Browser opened on NASA apod-page belonging to this background")
		return
	}

	if *login {
		today := a.Today()
		if downloaded, err := a.IsDownloaded(today); downloaded || err != nil {
			if err != nil {
				logger.Fatalf("Could not check whether today was downloaded, because: %v\n", err)
			}
			err := a.DisplayCurrent()
			if err != nil {
				logger.Fatalf("Today was already downloaded, but could not display the current wallpaper, because: %v\n", err)
			} else {
				logger.Printf("Displayed the current wallpaper, as today was already downloaded\n")
			}
			return
		}
		ok, err := a.Download(today)
		if err != nil {
			logger.Printf("An error occurred during todays (%s) image downloading: %v\n", today, err)
		}
		if !ok {
			logger.Printf("No new image today (%s) on APOD\n", today)

			apod.Notify(fmt.Sprintf("No new image today :-("))
			err := a.DisplayCurrent()
			if err != nil {
				logger.Fatalf("Could not display the current wallpaper, because: %v\n", err)
			} else {
				logger.Printf("Displayed the current wallpaper, as today had no new image\n")
			}
			return
		}
		err = a.SetWallpaper(apod.State{DateCode: today, Options: "fit"})
		if err != nil {
			logger.Fatalf("Could not set the wallpaper to %s, because: %v\n", today, err)
		} else {
			mesg := fmt.Sprintf("Wallpaper set to %s\n", today)
			apod.Notify(mesg)
			logger.Printf(mesg)
		}
		return
	}

	if *days > 0 {
		a.LoadRecentPast(*days)
	}

	if *jump != 0 {
		err := a.Jump(*jump)
		if err != nil {
			apod.Notify(err.Error())
			logger.Fatalf("Could not jump(%d): %v\n", *jump, err)
		}
		logger.Printf("Jump was successfull\n")
		return
	}

	if *mode {
		m, err := a.ToggleViewMode()
		if err != nil {
			logger.Fatalf("Could not mode viewing options: %v\n", err)
		} else {
			logger.Printf("Inversed the viewing option to: %s\n", m)
		}
		return
	}
}
