package apod

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	apodSite = "http://apod.nasa.gov/"
	format   = "060102"
)

var imageExpr = regexp.MustCompile(`<a href="(.*\.(jpg|gif|png))"`)

// ADate is an APOD date string
type ADate string

func NewADate(t time.Time) ADate {
	return ADate(t.Format(format))
}

func (d *ADate) String() string {
	return string(*d)
}

func (d *ADate) Back() *ADate {
	da := d.Date()
	if da == nil {
		return nil
	}
	p := NewADate(da.AddDate(0, 0, -1))
	return &p
}

func (d *ADate) Date() *time.Time {
	t, err := time.Parse(format, d.String())
	if err != nil {
		return nil
	}
	return &t
}

// APOD encapsulates communicating with apod.nasa.gov
type APOD struct {
	Client *http.Client
	Site   string
}

// NewAPOD constructs a new APOD object
func NewAPOD() *APOD {
	a := APOD{
		Client: http.DefaultClient,
		Site:   apodSite,
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
		return a.Site + "apod/" + m[1], nil
	}
	return "", nil
}

// Download fetches the url argument and stores the result in the path in the file argument
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
func (a *APOD) UrlForDate(isodate ADate) string {
	return fmt.Sprintf("%sapod/ap%s.html", a.Site, isodate.String())
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
