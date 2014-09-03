package apod

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/101loops/clock"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testDateString = "140121"
const configJSON = `{"WallpaperDir":"bar"}`

func TestMarshalConfig(t *testing.T) {
	cfg := Config{WallpaperDir: "bar"}
	text, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to write Config object to json: %v", err)
	}
	if string(text) != configJSON {
		t.Fatal("json has unexpected value")
	}
}

func TestUnmarshalConfig(t *testing.T) {
	var jsonBlob = []byte(configJSON)
	var cfg Config
	err := json.Unmarshal(jsonBlob, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WallpaperDir != "bar" {
		t.Fatal("Read unexpected value")
	}
}

func TestSetWallpaper(t *testing.T) {
	homeFile, err := spoofHome()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(homeFile)
	if err := MakeConfigDir(); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	logger := log.New(&buf, "", log.LstdFlags)
	apod := APOD{Config: cfg, Client: http.DefaultClient, Log: logger}
	ok, err := apod.Download(testDateString)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("21 januari 2014 does contain an image on APOD")
	}
	err = apod.SetWallpaper(State{DateCode: testDateString})
	if err != nil {
		t.Fatal(err)
	}
}

func spoofHome() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	homeFile := filepath.Join(cwd, "home")
	err = os.MkdirAll(homeFile, 0700)
	if err != nil {
		return "", err
	}
	err = os.Setenv("HOME", homeFile)
	return homeFile, err
}

func TestLoadConfig(t *testing.T) {
	homeFile, err := spoofHome()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(homeFile)
	if err := MakeConfigDir(); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cfg)
}

func TestToday(t *testing.T) {
	t0 := time.Date(2014, 1, 21, 0, 0, 0, 0, time.UTC)
	m := clock.NewMock()
	m.Set(t0)

	apod := APOD{Clock: m}
	if apod.Today() != testDateString {
		t.Errorf("Expected %v, got %v", testDateString, apod.Today())
	}
}

func TestUrlForDate(t *testing.T) {
	apod := APOD{}
	url := apod.UrlForDate(testDateString)
	if url != "http://apod.nasa.gov/apod/ap140121.html" {
		t.Errorf("Expected: http://apod.nasa.gov/apod/ap140121.html, got %s", url)
	}
}

func TestState(t *testing.T) {
	homeFile, err := spoofHome()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(homeFile)
	if err := MakeConfigDir(); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(stateFile())
	if err != nil {
		t.Fatal("could not create file foo")
	}
	_, err = f.WriteString(`{"DateCode":"140121","Options":"fit"}`)
	if err != nil {
		t.Fatal("could write to stateFile")
	}
	err = f.Close()
	if err != nil {
		t.Fatal("could close StateFile")
	}

	apod := APOD{}

	rv, err := apod.State()
	if err != nil {
		t.Fatalf("Error during call to State: %v\n", err)
	}
	if rv.DateCode != testDateString {
		t.Errorf("Expected 140121, got %v", rv)
	}
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
	apod := APOD{Client: &http.Client{Transport: apodRoundTrip{}}}
	_, url, err := apod.ContainsImage("http://apod.nasa.gov/apod/astropix.html")
	if err != nil {
		t.Fatal("could not load page")
	}
	if url != "http://apod.nasa.gov/apod/image/1407/nycsunset_tyson_768.jpg" {
		t.Fatalf("Expected http://apod.nasa.gov/apod/image/1407/nycsunset_tyson_768.jpg but got %v", url)
	}
}

func TestDownload(t *testing.T) {
	homeFile, err := spoofHome()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(homeFile)
	if err := MakeConfigDir(); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	apod := APOD{Config: cfg, Client: http.DefaultClient}
	_, err = apod.download("http://apod.nasa.gov/apod/image/1401/microsupermoon_sciarpetti_459.jpg", testDateString)
	if err != nil {
		t.Fatal("could not load page")
	}
	defer os.Remove(apod.fileName(testDateString))
	image := apod.fileName(testDateString)
	i, err := os.Open(image)
	if err != nil {
		t.Fatal(err)
	}
	info, err := i.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 181298 {
		t.Fatalf("Wrong size expected 181298")
	}
}

func prepareTest(t *testing.T) {
	err := os.MkdirAll("foo", 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir("foo")
	if err != nil {
		t.Fatal(err)
	}

	files := []string{"apod-img-000401", "apod-img-010401", "apod-img-020401"}
	for _, file := range files {
		f, err := os.Create(file)
		if err != nil {
			t.Fatal(err)
		}
		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}

	}
	err = os.Chdir("..")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloadedWallpapers(t *testing.T) {
	defer func() {
		os.RemoveAll("foo")
	}()
	prepareTest(t)
	cfg := Config{WallpaperDir: "foo"}
	apod := APOD{Config: cfg}

	files, err := apod.DownloadedWallpapers()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 3 {
		t.Fatal("Expected 3 files")
	}

}

func TestIndexOf(t *testing.T) {
	defer func() {
		os.RemoveAll("foo")
	}()
	apod := APOD{Config: Config{WallpaperDir: "foo"}}
	prepareTest(t)
	i, err := apod.IndexOf("010401")
	if err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatalf("Expected 1, got %d", i)
	}

}
func TestFileName(t *testing.T) {
	apod := APOD{Config: Config{WallpaperDir: "foo"}}
	expected := filepath.Join("foo", "apod-img-140121")
	got := apod.fileName(testDateString)
	if expected != got {
		t.Fatalf("Expected: %v, got %v", expected, got)
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

const imageResponseBase64 = `
begin-base64 644 IMAGE_RESPONSE
ICBIVFRQLzEuMSAyMDAgT0sKICBEYXRlOiBXZWQsIDAzIFNlcCAyMDE0IDE4
OjIyOjI3IEdNVAogIFNlcnZlcjogU3F1ZWVnaXQvMS4yLjUgKDNfc2lyKQog
IFZhcnk6ICoKICBYLVNlcnZlci1JUDogMjA5LjIwMi4yNDQuMTk1CiAgUDNQ
OiBwb2xpY3lyZWY9Imh0dHA6Ly93d3cubHljb3MuY29tL3czYy9wM3AueG1s
IiwgQ1A9IklEQyBEU1AgQ09SIENVUmEgQURNYSBERVZhIENVU2EgUFNBYSBJ
VkFhIENPTm8gT1VSIElORCBVTkkgU1RBIgogIENhY2hlLUNvbnRyb2w6IG1h
eC1hZ2U9NjA0ODAwCiAgRXhwaXJlczogV2VkLCAxMCBTZXAgMjAxNCAxODoy
MjoyNyBHTVQKICBMYXN0LU1vZGlmaWVkOiBUdWUsIDI4IFNlcCAxOTk5IDEw
OjMwOjIwIEdNVAogIEVUYWc6ICI4ODktMzdmMDk4YmMiCiAgQWNjZXB0LVJh
bmdlczogYnl0ZXMKICBDb250ZW50LUxlbmd0aDogMjE4NQogIENvbm5lY3Rp
b246IGNsb3NlCiAgQ29udGVudC1UeXBlOiBpbWFnZS9naWYKR0lGODlhMgAy
APcAAAAAAAgIAAgICBAIABAQABAQCBAQEBgYABgYCBgYEBgYGCEhCCEhECEh
GCEhISkpCCkpECkpISkpKTExITExKTExMTkxGDk5ITk5KTk5MTk5OUJCKUJC
MUJCOUpCQkpKOUpKQkpKSlJSKVJSSlJSUlpaUlpaWmNjMWNjUmNjWmNjY2tr
WmtrY2tra3Nzc3tza3t7Wnt7a3t7c3t7e4R7c4SEWoSEe4SEhISEjIyEhIyM
c4yMhIyMjIyMlJSMhJSMjJSMlJSUjJSUlJyUjJyUlJychJycjJyclJycnJyc
paWclKWcnKWcpaWlhKWlnKWlpaWlra2lnK2lpa2lra2tpa2tra2ttbWtpbWt
rbWttbW1nLW1pbW1rbW1tb21rb21tb21vb29pb29rb29tb29vb29xsa9tca9
vca9xsbGtcbGvcbGxsbGzs7Gvc7Gxs7Gzs7Ovc7Oxs7Ozs7O1tbOxtbOztbO
1tbWxtbWztbW1tbW3t7Wzt7W1t7W3t7ezt7e1t7e3ufe1ufe3ufe5+fn1ufn
3ufn5+fn7+/n3u/n5+/n7+/v3u/v5+/v7+/v9/fv5/fv7/fv9/f37/f39/f3
///39//3////9///////////////////////////////////////////////
////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////
/////////ywAAAAAMgAyAAAI/gABCBxIsKDBgwgTKlzIsKHDhxRIuJhB8QYP
GSUoPNxIUEGIGU+ydCHD5ssZMifZnMlyI0QBjgwLsICSBeUZOWzw4PyDM06e
P3jYkEkRACbCEEiekGQjByeeoHTY8PkyBk8eOVblfPlglKADFkSedMmpRg2Z
OHHcxFHjZYuXsj7j5OS5A0FXCkKeUJFSlkyUJ16oovVCh08hQ4gTJTLE0w0b
LhdgQgCCpEqXMWrUni3MGLGkRn8MPW1kqBGlRnx4qnmwEYMQJFyqVI3z1E2h
SZga6cZdKA6d37TVxMGEydBVMpEbOuDxBLCanGxo4/lDvLokSbTR0lEzZq2h
3H/+/rDxwnphABlLpJw9AzTtdDyJqhPnIzrN0zhpzKRR00hSIkDiGcFQCUg4
kUVUdDz1Fh8MYjIJbqTd1luCeODnRht0JNLIfzydoJADRFCRxRpAHZJIHFwI
F4duvz34xyOTIOLbH3ughQcfdISm4SFyjHFAQjN18YVUgEyyGB6N9EajkpP8
8aCMdOxB4x6/hVbahjzVgJADQlBBBh505PEgJhoWUlV4NL74nYON7FEIIn78
UYhhjcBIJiBBlVfQCHqRhIchkYzZCFt/IlKInA/amdskoIW3h2GGSEJcIjyh
cJANTywB2HROEjeJcHg8Qtoegcjn4COPGJLqI35MQuek/nnEUYRBDfDgBA9K
DGlIIpNEKokaYBaCyR9xNmLdHqWtGlqMhRjr3x9ufPEjQRkIscQOR1zBx6T0
TRJsIZK8eZ18kT44SaqSpHobJs/icYYIezYXRLbfJTnGodPRURpukuZmqqcO
Oiisg6LJcUYMBcmgFBFKcMHoJGsluMd0/Ur62biS4PbvucQZgi8bTRAUAA4h
kXEFdYW44cUYbci5pKQQSmJIk6VqLN8k/SaJmBtXFCVQATwc8YRPjQAShyFt
YIjHYYEM7KnGDxq7cXWMFqJTGD4DYABzXLDRXiIVggmUqI/0CzBxkuzxSHFj
mprqHwluQcBABdzQnJBSuOEx/h1WA4UHIvH9S7AfjZxrs7mYHPoHWltkDcAN
Ql/hRIpH/ynJU2NvjAgiGV/sYMzhLo2HG3hcUdAMSGQhxRVXCLctcU8F0p/g
jzDy9KmF48x0eG7kEURBLFQmBRetsyH1JFabLbjgjSDS3yOhhYcWIAgTFMIS
ZMhB/FsDfxbe8v92vwdu0PftNSAcFBSBEFWw0cUVXMQR+GfN2jx1k9Q12l9n
7yUyx0siu0F6hiccnAElXMuTWnX+0B8pSakzjAoNII5wkBII4QhSoMJ0IhEJ
PKjhD5QAX3XqtAdE7KF5icPDuBqBB0AAwgQHccAOiFAFLnQhPJMgQxrioLzl
NUlK/of6zmFukzFAJEIOCkBIC4SQhTi0zzBkuFcPpxaeHIGLXZ/64IbyMIOE
RIAHVTgDe0TTGREGrElR4gPZWOiGDDXiEGQAIEJM8ISbuNAQR4sUcShhP0/J
jA5Ju4OqFsOHOPBhizBUSABu0L6fyOEPfIhPCO93rt8UhjTGOpSR9ICEhuAl
C3LIg9cO8QdemQpCh0OE7ZBHh9NoCBBgSGJDQICELvQkPEa0mbOuw69FMSox
hQCEHtiQgY1woEskkYMLFTMmIzEqPjjDYuGmaZwzOAAmHuAB9kL5k0E0IlAa
+kMkeKUhSVDCnEYqmhzI4IGuZAB1YGBDHvIAiEMwBhC7c0rMrqapmEOIcgnX
7AoACtCC1CXTKnl4D4AoZYhD4GkQwyRDDOQoUABQwAZJoMJYosPNPJwBEPJE
qEqC0ICKHgQDYNHoSMBAEjKQwQ1nGMkTVlBMkyakABhIAQ2SkqknQOEIOljB
BBxn06Ia9ahITSpBAgIAOw==
====`
