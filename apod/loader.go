package apod

import (
	"fmt"
)

type Loader struct {
	APOD   *APOD
	Config *config
	Notifier
	logger
}

// Download downloads the image from apod.nasa.gov for the given date.
func (l *Loader) Download(isodate ADate) (bool, error) {
	if downloaded, _ := l.Config.IsDownloaded(isodate); downloaded {
		return true, nil
	}
	pageURL := l.APOD.UrlForDate(isodate)
	imgURL, err := l.APOD.ContainsImage(pageURL)
	if err != nil {
		return false, err
	}
	if imgURL == "" {
		return false, nil
	}
	l.Notify(fmt.Sprintf("Downloading APOD-image for: %s", isodate))
	file := l.Config.fileName(isodate)
	err = l.APOD.Download(file, imgURL)
	if err != nil {
		return true, err
	}
	l.Printf("Successfully downloaded %s to %q\n", isodate, file)
	return true, nil
}

// LoadPeriod loads images from apod.nasa.gov to the wallpaper directory, for a number of days back
func (l *Loader) LoadPeriod(from ADate, days int) error {
	for _, isodate := range l.days(from, days) {
		_, err := l.Download(isodate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) days(from ADate, days int) []ADate {
	var dates []ADate
	for i := 1; i < days+1; i++ {
		dates = append(dates, NewADate(from.Date().AddDate(0, 0, -i)))
	}
	return dates
}
