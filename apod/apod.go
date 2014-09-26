package apod

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

const (
	apodBase = "http://apod.nasa.gov/apod/"
	format   = "060102"
)

var imageExpr = regexp.MustCompile(`<a href="(.*\.(jpg|gif|png))"`)

// APOD encapsulates communicating with apod.nasa.gov
type APOD struct {
	Client *http.Client
}

// NewAPOD constructs a new APOD object
func NewAPOD() *APOD {
	a := APOD{
		Client: http.DefaultClient,
	}
	return &a
}

// ContainsImage parses an APOD page for a linked image, returns image URL if successful (maybe empty)  or an error
func (a *APOD) ContainsImage(url string) (string, error) {
	content, err := a.loadPage(url)
	if err != nil {
		return "", err
	}
	m := imageExpr.FindStringSubmatch(content)
	if m != nil && m[1] != "" {
		return apodBase + m[1], nil
	}
	return "", nil
}

func (a *APOD) Download(file, url string) error {
	output, err := os.Create(file)
	if err != nil {
		return err
	}
	defer output.Close()
	resp, err := a.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	return err
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// UrlForDate returns the URL for the APOD page for the given ISO date.
func (a *APOD) UrlForDate(isodate string) string {
	return fmt.Sprintf("http://apod.nasa.gov/apod/ap%s.html", isodate)
}

func (a *APOD) loadPage(url string) (string, error) {
	resp, err := a.Client.Get(url)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
