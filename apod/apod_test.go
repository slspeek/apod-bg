package apod

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"testing"
	"time"
)

func testApod() *APOD {
	a := NewAPOD()
	a.Site = "http://localhost:8765/"
	return a
}

func TestADateString(t *testing.T) {
	a := ADate("140121")
	assert.Equal(t, "140121", a.String())
}

func TestADateDate(t *testing.T) {
	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	a := ADate("140121")
	assert.Equal(t, &t0, a.Date())
}

func TestContainsImageWithServer(t *testing.T) {
	go func() {
		http.ListenAndServe(":8765", http.FileServer(http.Dir("../testdata/apod.nasa.gov/")))
	}()
	testHome := setupTestHome(t)
	defer cleanUp(t, testHome)
	a := testApod()
	url, err := a.ContainsImage(a.UrlForDate(testDateSeptember))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a.Site+"apod/image/1409/m8_chua_2500.jpg", url)
}

func TestCollectTestData(t *testing.T) {
	t.Skip()
	resp, err := http.Get("http://timbeauchamp.tripod.com/moon/moon15.gif")
	if err != nil {
		t.Fatal(err)
	}

	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile("../testdata/moon15.gif.response", dump, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUrlForDate(t *testing.T) {
	apod := APOD{Site: apodSite}
	url := apod.UrlForDate(testDateString)
	assert.Equal(t, "http://apod.nasa.gov/apod/ap140121.html", url)
}

type testRoundTrip struct{}

func (l testRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.String() == "http://apod.nasa.gov/apod/ap140121.html" {
		return apodRoundTrip{}.RoundTrip(r)
	} else {
		return imageRoundTrip{}.RoundTrip(r)
	}
}

type imageRoundTrip struct{}

func (l imageRoundTrip) RoundTrip(*http.Request) (*http.Response, error) {
	f, err := os.Open("../testdata/moon15.gif.response")
	if err != nil {
		panic(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(f), nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type apodRoundTrip struct{}

func (l apodRoundTrip) RoundTrip(*http.Request) (*http.Response, error) {
	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(apodResponse)), nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func TestContainsImage(t *testing.T) {
	apod := APOD{Client: &http.Client{Transport: apodRoundTrip{}}, Site: apodSite}
	url, err := apod.ContainsImage("http://apod.nasa.gov/apod/astropix.html")
	if err != nil {
		t.Fatal("could not load page")
	}
	if url != "http://apod.nasa.gov/apod/image/1407/nycsunset_tyson_768.jpg" {
		t.Fatalf("Expected http://apod.nasa.gov/apod/image/1407/nycsunset_tyson_768.jpg but got %v", url)
	}
}

func TestLoadPage(t *testing.T) {
	apod := APOD{Client: &http.Client{Transport: apodRoundTrip{}}}
	page, err := apod.loadPage("http://apod.nasa.gov/apod/astropix.html")
	if err != nil {
		t.Fatalf("Error loading page: %v", err)
	}
	if len(page) != 4956 {
		t.Fatalf("Wrong length wanted 4956, got: %d", len(page))
	}
}

var apodResponse = `HTTP/1.1 200 OK
Date: Mon, 07 Jul 2014 12:06:30 GMT
Server: WebServer/1.0
Accept-Ranges: bytes
Content-Length: 4956
Keep-Alive: timeout=5, max=100
Connection: Keep-Alive
Content-Type: text/html; charset=ISO-8859-1

<html>
<head>
<title> APOD: 2014 July 6 - Manhattanhenge: A New York City Sunset   
</title> 
<!-- gsfc meta tags -->
<meta name="orgcode" content="661">
<meta name="rno" content="phillip.a.newman">
<meta name="content-owner" content="Jerry.T.Bonnell.1">
<meta name="webmaster" content="Stephen.F.Fantasia.1">
<meta name="description" content="A different astronomy and space science
related image is featured each day, along with a brief explanation.">
<!-- -->
<meta name="keywords" content="Manhattan, Stonehenge, sunset">
<script id="_fed_an_js_tag" type="text/javascript"
src="js/federated-analytics.all.min.js?agency=NASA"></script> 
</head>
<body BGCOLOR="#F4F4FF" text="#000000" link="#0000FF" vlink="#7F0F9F"
alink="#FF0000">

<center>
<h1> Astronomy Picture of the Day </h1>
<p>

<a href="archivepix.html">Discover the cosmos!</a>
Each day a different image or photograph of our fascinating universe is
featured, along with a brief explanation written by a professional astronomer.
<p>

2014 July 6 
<br> 
<a href="image/1407/nycsunset_tyson_768.jpg">
<IMG SRC="image/1407/nycsunset_tyson_960.jpg"
alt="See Explanation.  Clicking on the picture will download
 the highest resolution version available."></a>
</center>

<center>
<b> Manhattanhenge: A New York City Sunset </b> <br> 
<b> Image Credit & Copyright: </b> 
<a href="http://research.amnh.org/users/tyson/">Neil deGrasse
Tyson</a> 
(<a href="http://amnh.org/rose/">AMNH</a>)
</center> <p> 

<b> Explanation: </b> 
This coming Saturday, if it is clear, well placed New Yorkers can 
<a href=
"http://www.amnh.org/our-research/hayden-planetarium/resources/manhattanhenge">go outside at sunset</a> and watch their city act like a modern version of 
<a href="ap990912.html">Stonehenge</a>.  

<a href="http://en.wikipedia.org/wiki/Manhattan"
>Manhattan</a>'s streets will flood dramatically with 
sunlight just as the Sun sets precisely at each street's western end.

Usually, the <a href="http://en.wikipedia.org/wiki/List_of_tallest_buildings_in_the_world#Tallest_skyscrapers_in_the_world">tall buildings</a> 
that line the gridded streets of 
<a href="http://en.wikipedia.org/wiki/History_of_New_York_City"
>New York City</a>'s tallest borough will hide the setting Sun.  

This effect makes <a href="ap131104.html">Manhattan</a> 
a type of modern 
<a href="http://www.britannia.com/history/h7.html">Stonehenge</a>, 
although only aligned to about 30 
<a href="http://aleph0.clarku.edu/~djoyce/java/trig/angle.html"
>degrees</a> east of north.  

Were <a href=
"http://www.brainpickings.org/index.php/2012/01/17/the-greatest-grid/">Manhattan's road grid</a> perfectly aligned to east and west, 
today's effect would occur on the 
<a href="http://en.wikipedia.org/wiki/Equinox">Vernal</a> and 

<a href="ap030923.html">Autumnal Equinox</a>, 
March 21 and September 21, the only two days that the 
<a href="ap040320.html">Sun rises and sets due east and west</a>.  

<a href=
"http://www.amnh.org/layout/set/plain/content/view/popup/53923/(baseNodeID)/3278"
>Pictured above</a> in this horizontally stretched image,
the Sun sets down
<a href="http://en.wikipedia.org/wiki/34th_Street_(Manhattan)"
>34th Street</a> as
viewed from 
<a href="http://en.wikipedia.org/wiki/Park_Avenue">Park Avenue</a>.

If Saturday's sunset is hidden by clouds <a href=
"http://img2.wikia.nocookie.net/__cb20130624220704/animaljam/images/7/7a/Bear-sitting-picnic-table.jpg"
>do not despair</a> -- the same thing happens twice each year:  
in late May and mid July.  

On none of these occasions, however, should you ever 
look directly at the Sun.


<p> <center> 
<b> Tomorrow's picture: </b><a href="ap140707.html">three black holes</a>

<p> <hr>
<a href="ap140705.html">&lt;</a>
| <a href="archivepix.html">Archive</a>
| <a href="lib/aptree.html">Index</a>
| <a href="http://antwrp.gsfc.nasa.gov/cgi-bin/apod/apod_search">Search</a>
| <a href="calendar/allyears.html">Calendar</a>
| <a href="/apod.rss">RSS</a>
| <a href="lib/edlinks.html">Education</a>
| <a href="lib/about_apod.html">About APOD</a>
| <a href=
"http://asterisk.apod.com/discuss_apod.php?date=140706">Discuss</a>
| <a href="ap140707.html">&gt;</a>

<hr><p>
<b> Authors & editors: </b>
<a href="http://www.phy.mtu.edu/faculty/Nemiroff.html">Robert Nemiroff</a>
(<a href="http://www.phy.mtu.edu/">MTU</a>) &
<a href="http://antwrp.gsfc.nasa.gov/htmltest/jbonnell/www/bonnell.html"
>Jerry Bonnell</a> (<a href="http://www.astro.umd.edu/">UMCP</a>)<br>
<b>NASA Official: </b> Phillip Newman
<a href="lib/about_apod.html#srapply">Specific rights apply</a>.<br>
<a href="http://www.nasa.gov/about/highlights/HP_Privacy.html">NASA Web
Privacy Policy and Important Notices</a><br>
<b>A service of:</b>
<a href="http://astrophysics.gsfc.nasa.gov/">ASD</a> at
<a href="http://www.nasa.gov/">NASA</a> /
<a href="http://www.nasa.gov/centers/goddard/">GSFC</a>
<br><b>&</b> <a href="http://www.mtu.edu/">Michigan Tech. U.</a><br>
</center>
</body>
</html>
`
