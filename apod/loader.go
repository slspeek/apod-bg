package apod

import (
	"fmt"
	"time"
)

type Loader struct {
	APOD    *APOD
	storage *Storage
	Notifier
	logger
}

// Download downloads the image from apod.nasa.gov for the given date.
func (l *Loader) Download(isodate string) (bool, error) {
	if downloaded, _ := l.storage.IsDownloaded(isodate); downloaded {
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
	file := l.storage.fileName(isodate)
	err = l.APOD.Download(file, imgURL)
	if err != nil {
		return true, err
	}
	l.Printf("Successfully downloaded %s to %q\n", isodate, file)
	return true, nil
}

// LoadPeriod loads images from apod.nasa.gov to the wallpaper directory, for a number of days back
func (l *Loader) LoadPeriod(from time.Time, days int) error {
	for _, isodate := range l.days(from, days) {
		_, err := l.Download(isodate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) days(from time.Time, days int) []string {
	var dates []string
	for i := 1; i < days+1; i++ {
		dates = append(dates, from.AddDate(0, 0, -i).Format(format))
	}
	return dates
}
