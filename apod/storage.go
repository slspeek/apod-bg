package apod

import (
	"fmt"
	"os"
	"sort"
)

const imgPrefix = "apod-img-"

func stripPrefix(s string) string {
	return s[len(imgPrefix):]
}

type Storage struct {
	Config *config
}

// IsDownloaded checks whether an image file is downloaded for a given date.
func (c *config) IsDownloaded(isodate ADate) (bool, error) {
	file := c.fileName(isodate)
	fileExists, err := exists(file)
	if err != nil {
		return false, err
	}
	return fileExists, nil
}

func (s *Storage) DownloadedWallpapers() ([]ADate, error) {
	dir, err := os.Open(s.Config.WallpaperDir)
	if err != nil {
		return nil, err
	}
	files, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	dates := []ADate{}
	for _, f := range files {
		dates = append(dates, ADate(stripPrefix(f)))
	}
	return dates, nil
}

// IndexOf returns the index of an image in the wallpaper directory.
func (s *Storage) IndexOf(isodate ADate) (int, error) {
	all, err := s.DownloadedWallpapers()
	if err != nil {
		return 0, err
	}
	for i, elem := range all {
		if elem == isodate {
			return i, nil
		}
	}
	return 0, fmt.Errorf("%s was not found", isodate)
}
