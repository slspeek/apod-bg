package main

import (
	"flag"
	"fmt"
	"github.com/101loops/clock"
	"github.com/slspeek/apod-bg/apod"
	"net/http"
)

var (
	info   = flag.Bool("info", false, "open the APOD-page on the current background")
	login  = flag.Bool("login", false, "Do the login procedure: fetch today and show it")
	fetch  = flag.Bool("fetch", false, "do a batch fetch from the past")
	days   = flag.Int("days", 14, "days to go back downloading")
	jump   = flag.Int("jump", 0, "number of images to jump")
	config = flag.String("config", "", "configuration to write out")
)

func main() {
	flag.Parse()
	fmt.Println("Welcome to APOD Background")
	cfg, err := apod.LoadConfig()
	if err != nil {
		fmt.Printf("Could not load the configuration, because: %v\n", err)
	}
	a := apod.APOD{Config: cfg, Clock: clock.New(), Client: http.DefaultClient}

	if *info {
		err = a.OpenAPODOnBackground()
		if err != nil {
			fmt.Printf("Could not open the APOD page on background now showing, because: %v\n", err)
		}
		fmt.Println("Opened the default browser on the APOD-page related to the current background image")
		return
	}
	if *login {
		today := a.Today()
		ok, err := a.Download(today)
		if err != nil {
			fmt.Printf("An error occurred during todays image downloading: %v\n", err)
		}
		if !ok {
			fmt.Println("No new image today on APOD")
			a.DisplayCurrent()
			return
		}
		a.SetWallpaper(today)
	}
	if *fetch {
		a.LoadRecentPast(*days)
	}
	if *jump != 0 {
		err = a.Jump(*jump)
		if err != nil {
			fmt.Printf("Could not jump(%d): %v\n", *jump, err)
			return
		}
		fmt.Println("Jump was successfull\n")
		return
	}
	if *config != "" {
		script := ""
		switch *config {
		case "barewm":
			script = apod.SetWallpaperScriptBareWM
		case "lxde":
			script = apod.SetWallpaperScriptLXDE
		}
		err = apod.WriteConfig(script)
		if err != nil {
			fmt.Printf("Could not write the configuration, because: %v\n", err)
		}
	}
}
