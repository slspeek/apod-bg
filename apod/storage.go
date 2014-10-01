package apod

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Storage struct {
	Config *config
}

// IsDownloaded checks whether an image file is downloaded for a given date.
func (s *Storage) IsDownloaded(isodate string) (bool, error) {
	file := s.fileName(isodate)
	fileExists, err := exists(file)
	if err != nil {
		return false, err
	}
	return fileExists, nil
}

func (s *Storage) DownloadedWallpapers() ([]string, error) {
	dir, err := os.Open(s.Config.WallpaperDir)
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

// IndexOf returns the index of an image in the wallpaper directory.
func (s *Storage) IndexOf(isodate string) (int, error) {
	target := s.fileBaseName(isodate)
	all, err := s.DownloadedWallpapers()
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

func (s *Storage) fileName(isodate string) string {
	return filepath.Join(s.Config.WallpaperDir, s.fileBaseName(isodate))
}

func (s *Storage) fileBaseName(isodate string) string {
	return fmt.Sprintf(imgPrefix+"%s", isodate)
}
