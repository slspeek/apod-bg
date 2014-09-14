package apod

import (
	"time"
)

type Loader struct {
	a       *APOD
	storage *Storage
}

// Download downloads the image from apod.nasa.gov for the given date.
func (l *Loader) Download(isodate string) (bool, error) {
	if downloaded, _ := l.storage.IsDownloaded(isodate); downloaded {
		return true, nil
	}
	pageURL := l.a.UrlForDate(isodate)
	imgURL, err := l.a.ContainsImage(pageURL)
	if err != nil {
		return false, err
	}
	if imgURL == "" {
		return false, nil
	}
	//Notify(fmt.Sprintf("Downloading todays APOD-image"))
	file := l.storage.fileName(isodate)
	err = l.a.Download(file, imgURL)
	if err != nil {
		return true, err
	}
	//a.Log.Printf("Successfully downloaded %s to %q\n", isodate, file)
	return true, nil
}

// LoadRecentPast loads images from apod.nasa.gov to the wallpaper directory, for a set number of days back
func (l *Loader) LoadRecentPast(from time.Time, days int) {
	for _, isodate := range l.recentPast(from, days) {
		l.Download(isodate)
	}
}

func (l *Loader) recentPast(from time.Time, days int) []string {
	var dates []string
	for i := 1; i < days+1; i++ {
		dates = append(dates, from.AddDate(0, 0, -i).Format(format))
	}
	return dates
}
