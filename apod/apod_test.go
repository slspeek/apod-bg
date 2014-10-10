package apod

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"testing"
)

const testAPODSite = "http://localhost:8765/"

func testAPOD() *APOD {
	a := NewAPOD()
	a.Site = testAPODSite
	return a
}

func TestADateString(t *testing.T) {
	a := ADate("140121")
	assert.Equal(t, "140121", a.String())
}

func TestADateDate(t *testing.T) {
	a := ADate("140121")
	assert.Equal(t, &testDate, a.Date())
}

func TestNewADate(t *testing.T) {
	a := NewADate(testDate)
	assert.Equal(t, &testDate, a.Date())
	assert.Equal(t, "140121", a.String())
}

func TestADateBack(t *testing.T) {
	a := ADate("140101")
	b := a.Back()
	assert.Equal(t, "131231", b.String())
}

func TestContainsImageWithServer(t *testing.T) {
	go func() {
		http.ListenAndServe(":8765", http.FileServer(http.Dir("../testdata/apod.nasa.gov/")))
	}()
	testHome := setupTestHome(t)
	defer cleanUp(t, testHome)
	a := testAPOD()
	url, err := a.ContainsImage(a.UrlForDate(testDateSeptember))
	assert.NoError(t, err)
	assert.Equal(t, a.Site+"apod/image/1409/m8_chua_2500.jpg", url)
}

func TestCollectTestData(t *testing.T) {
	t.Skip()
	resp, err := http.Get("http://timbeauchamp.tripod.com/moon/moon15.gif")
	assert.NoError(t, err)

	dump, err := httputil.DumpResponse(resp, true)
	assert.NoError(t, err)

	err = ioutil.WriteFile("../testdata/moon15.gif.response", dump, 0644)
	assert.NoError(t, err)
}

func TestUrlForDate(t *testing.T) {
	apod := NewAPOD()
	url := apod.UrlForDate(testDateString)
	assert.Equal(t, "http://apod.nasa.gov/apod/ap140121.html", url)
}

func TestContainsImage(t *testing.T) {
	apod := testAPOD()
	url, err := apod.ContainsImage(testAPODSite + "apod/ap140921.html")
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8765/apod/image/1409/saturnequinox_cassini_7227.jpg", url)
}

func TestLoadPage(t *testing.T) {
	a := testAPOD()
	page, err := a.loadPage(testAPODSite + "apod/ap140921.html")
	assert.NoError(t, err)
	assert.Equal(t, 5069, len(page))
}
