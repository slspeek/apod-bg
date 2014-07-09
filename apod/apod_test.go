package apod

import (
	"bufio"
	"github.com/101loops/clock"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testDateString = "140121"

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

func TestNowShowing(t *testing.T) {
	f, err := os.Create("foo")
	if err != nil {
		t.Fatal("could not create file foo")
	}
	defer os.Remove("foo")
	_, err = f.WriteString(`feh  --bg-max '140121-apod.jpg'`)
	if err != nil {
		t.Fatal("could write to file  foo")
	}
	err = f.Close()
	if err != nil {
		t.Fatal("could close file foo")
	}
	cfg := Config{StateFile: "foo"}

	apod := APOD{Config: cfg}

	rv, err := apod.NowShowing()
	if err != nil {
		t.Fatalf("Error during call to NowShowing")
	}
	if rv != testDateString {
		t.Errorf("Expected 140121, got %v", rv)
	}
}

type m13response struct{}

func (l m13response) RoundTrip(*http.Request) (*http.Response, error) {
	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(m13imageResponse)), nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type loopback struct{}

func (l loopback) RoundTrip(*http.Request) (*http.Response, error) {
	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(apodResponse)), nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func TestContainsImage(t *testing.T) {
	apod := APOD{Client: &http.Client{Transport: loopback{}}}
	yes, url, err := apod.ContainsImage("http://apod.nasa.gov/apod/astropix.html")
	if err != nil {
		t.Fatal("could not load page")
	}
	t.Logf("Found: %v; URL: %s", yes, url)
	if url != "http://apod.nasa.gov/apod/image/1407/nycsunset_tyson_768.jpg" {
		t.Fatalf("Expected http://apod.nasa.gov/apod/image/1407/nycsunset_tyson_768.jpg but got %v", url)
	}
}

func TestDownload(t *testing.T) {
	apod := APOD{Client: http.DefaultClient}
	err := apod.Download("http://apod.nasa.gov/apod/image/1401/microsupermoon_sciarpetti_459.jpg", testDateString)
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

func TestDownloadedWallpapers(t *testing.T) {
	defer func() {
		os.RemoveAll("foo")
	}()
	err := os.MkdirAll("foo", 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir("foo")
	if err != nil {
		t.Fatal(err)
	}

	one, err := os.Create("bar")
	if err != nil {
		t.Fatal(err)
	}
	one.Close()
	two, err := os.Create("baz")
	if err != nil {
		t.Fatal(err)
	}
	two.Close()
	three, err := os.Create("foobar")
	if err != nil {
		t.Fatal(err)
	}
	three.Close()
	err = os.Chdir("..")
	if err != nil {
		t.Fatal(err)
	}

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
func TestFileName(t *testing.T) {
	apod := APOD{Config: Config{WallpaperDir: "foo"}}
	expected := filepath.Join("foo", "apod-img-140121")
	got := apod.fileName(testDateString)
	if expected != got {
		t.Fatalf("Expected: %v, got %v", expected, got)
	}
}

func TestLoadPage(t *testing.T) {
	cfg := Config{StateFile: "foo"}
	apod := APOD{Config: cfg, Client: &http.Client{Transport: loopback{}}}
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

var m13imageResponse = `HTTP/1.1 200 OK
  Date: Wed, 09 Jul 2014 10:29:13 GMT
  Server: Apache/2.2.24 (Unix) mod_ssl/2.2.24 OpenSSL/1.0.0-fips mod_bwlimited/1.4 PHP/5.3.23
  Last-Modified: Thu, 31 Oct 2013 21:19:04 GMT
  ETag: "47dfd6-6232-4ea0fffd6a2a8"
  Accept-Ranges: bytes
  Content-Length: 25138
  Keep-Alive: timeout=5, max=100
  Connection: Keep-Alive
  Content-Type: image/png

iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAYAAABccqhmAAAgAElEQVR42u29WZBc15nfidoLVdhRRAGFwg4UUARQAGrf931H7QtqrwJFUVSTIiVKFEVSorioJZLiIlEKz8yjwx0xtqMd82Z7/OB5mO7whF/GfrCeZjzunrGjPfbowd1PZ87vu3mzbmZlZuVy7817M4sRh6glK/Pce8/5n2/5f//v2LFjdYqRk1OvenqG1blzLcr8mZdHa+ugKi9vC36fm1uvnv1yS52x/MyJMfeTNXW+siPu13d1DatTp5p9cU+9PM6ebVZtbYOen2dZWasv5skoLm7Q/wa+uXChVbW3D6VxQrV6vKBHHpOyjEI9bgV+fyzquPaoRk39YEV/fUaPHMvveL+rts1z7ZfbquT0lRjzNF97XY+CCK+5p8djh+d52L3kNaej3Mu8wO8fO34v45unMQoKalReXl5gPoxTHptnpN/z/SUX55nMCHzR2DigKivb0ziRR/oUL1UPHjxQvb29amRkRHV2dupTvjxw0+5rKyVHjY+PHxj3799X/S+Nqhc7G1R+fqGqra2Vv+/v71fXrl0L/P3dlOeIlbT72z39GSdjzPOBHrfl4VZWVmqrqkeNjo7Kv2fOnA/87ryj8zz8XjLHfP3MG0Pu4/DwsAUAnJ5jvPM0X3tSlZSUyGuamppk/t6a5xn9+4eqr69Pnjf/3r17V6+ZXP27Gy7NM0kAKC1tVAMDo2JGpxMAzI11+vRpVVRUJDeIm1laWqp/f0VQMzc3NzgKCgrkZl6qvKQ2v9hSRSXFqqamRrW0tKjCwkLtzpyTRc2/xomXosl0skmtf7YdxzyL9OK4qO/pgCwSXsNrS0pOBk6MHEfnGd+9zBEAuHr1asg9NQCg1oU5xjvPusA9y1MNDQ3a9WsNAECOx+Z5Rn9mmYAU6/Ls2bPy/C9fvqx/V+7SPJMEgPv3e9S9e91pNkUeB8yoM4FRJhuek7OiokJ/fVGPJ/Iwurr69WItlhMWRH6x80VtAUyJhQAgGDe1MHjTGfsnWwp+aEWbWnh/PY55HlMdHR3qypUrFnMvN2C23nF8nvHdSwMAzDkyp2PHSmR+prXl7BwTeeYF6uLFiwIAt2/ftgCAl+Z5N/C7XPkdm5zfG1bCeZfmmcTIz69XQ0NjgYCAVwIUIPsJQdqxsTF14sQJ/X2l/O706WbV3NwtN7m5uVlVVVWpqbem1bWaJkFjTNn8/PyAqV2krl+/rtra2gILJrV5VdztVhNvrBwyz4tykjIPFiumHicBp0dubr4sEqfnGd+9zBMAwFRlcC/Pnz8f+PxKl+cYa57l2vfPV93d3XK6hgLATQ/Ns07meu9etRocHJTfMVcD/M+lYZ5xjuvXO9STJ30e2/xn9GbJE3OJjWPcxPvy+zt3uvRCvaKOHz8uN/nitYvq2S/X9etviHnGTTZe/1BONKyErq6uwM9Sm9ut+gHV/3z+kHnekIXKPDBXWSjMFZ8RsGLjOT3P+O5lqbpwoVzmwiJmsWLSnjx5Uvxt9+YYa55GLKW6ujpw747ped6xAMA1j8zzfhAA8PN53pj+Q0ND6swZrILjLs8zgUGayjupP+MG5+QY/l5dXZ18bZhfxmtIsXCS3rlzRx5C/WS9al2YFD8xFGUxZYttRdkHPSOqbXn6kHnek03PPDBbc3IKZBHja+MWMA+n5xnfvbwd8F0vqVOnymURYhEYG63IpTkeNs9LAkic/kZ84pjFAjgmgUBvzNPqKtwPWAQ56uHDhwETv8DFeSYBAN7Z/GclalpfXy832Yignha/n98XFDToh98iDx/fH/915aNVVXatI4I/WGy7n1U/OaHqBMVjzfOhzA8zEP8vJ6fwAAA4Pc/D7+XjQJoSALis53FN5sxr7927p0698ILKL8h3eI7xzLNcgm3MAzeKQeDMjLIDAM7fy3jm+Siw6S8GAn6Vcj8BAMNKKHBpnkmMa9c6PAIAD+QmkSbhZAcpzci0EaC6oioq2rVJVRaMoFbeq1Rz7y5bHpJxUzkhzEisHZHWS3e61cBLC2rl403Vuz2jcnJjz5NAEGYrLgDBoOLiYosLUODYPOO/lxXary7RZmqlzI05AlAz6zNq5+sdtfe7PbX7za4aXB10cI7xzPOs+P9YVOYgtcZr+dqdexnf/czPL1U3btwMZgEuXbokG76srCwwDzfmmcTIy6v3DACYvnP4uHHjhmysR4961dUX76jFnyyqtU/W1NxP5lTt2JjlPUJzrZwYqeZar9zvVc9/91w9//3+6FzujDlPM3DFA+chMw+4CkYQ8JIj80zsXhoLFoBifswB62Tjsw01+iejqqjkgbZ2OgUEnJtjPPOsDOTYq/WokrgFrp/hAnBq1rhwL+O7nwAq88LvN3kCxAEMgLjj0jyTGJcvt4tpnX4AeBiT6cfGGluYktPJuhmn31oN88HsZVtN/XAl5PMYO7/dESsg2jyxRk6cKA+YidY04IWApeI0K+ywe3lZNlPo5x+TDX/lfmOAWv1EgK/0zAkHmWuHP/PQ11/W1gon7HGXmYDx3M+TB+4n6b5Hj54oMm1eZALm5+t9DwNweHhMnT5d5gFqaPRx8mSTWn772YHNyDh5vjX0wgob9CnWpI6falKnylrV6QttEidgXL7Xoy5Xd6vrj/vUzbp+dad5QFV3DKv73SPq8dCYeqzvRePMpIz2ladq64udiJ9ZVNoYc77V1T2SsfADJ/z27S7hgax8uKkWfrquLt3tUsOvLKrd3+ypYznemisu66NHfb64r4ympgF18WKbZ+bDYV9R0aYaGvpl3x8zkKBeI36BB6ih0cfNm51q8Z3IAMAm3fh8W+19Y3y//dWufA9vf+nDDbX0wYaa0eDBGH99WY29tqwGX16QlF7P1pzqePZUtS8/DW58AQI9XuwaVpNvHbQAlj/eOHS+dRpciFn4YZHeutUpgHX6Yqta/WRTTv7NX++oaw+9t9HOn2+RQjC/AABp9pqa3rTO4fjxRu2qdKpmfdgNDY2q+vp+deVKR8DyD74wMjUUX/bOnc60UxlB0ubJiYgAsPnVjmzuh32j2jS393MLSxrV7Dtrwc9a/3RblV09PHDa0TEkpCV/AECXBoBuX8y1uNigrfsFAEr0+unvH3H9c7GYq6q6ZR1C9IPrgyWSmxulGCicGmqc7CX6jaq1idgUksbIyytyNY1BoBJzBcTq2Z7dP/m/3BF23omzLeqmRrWVjzYtRJ3QkVKdgzaDz19pV+U3u8S9iOdvmK/h+x0BgN3DXAt+mW9397DjhwH349KlNvXwYa/q7R2R0v4XX+wRiyknJ45qwESooXfvPpGoqFtEBur+MV+Cp0Bpk1rUZn1VmCnIyUyQ8MKNzmDp7uORMTX7k2fi25q5fNJ5gAcn+6WqfT+95EyzAMjar7bUxmc7qmtjJvi7UxdaxXXY/HxH3gvAiWo1aJAAdf2yQHGvWCx+ma+frCuz1oY4i93vi84E4N3SMqgP5zHZIzzL0kPiU1EAIDFqaE5OqYXKmCNmDmYGvkVRUaPtNxA/1fqz+ffW1Z2mUAB40Deinn/zXEp28eUBgOFXF1Ue0c4ACt5uGJDgIK7CPb2Q2Oz5BQ1S6vv07VUJ/BUUN8jfXAwE8XgtRUCAR64+1fk5AIJVEFm8oiXN2gqZDQC1tX2K7JVf5vvCC622xC3w5SnZf/y4T9wgTnn2BqZ98tZm8It9aijFC4lSQ0tKmiRC29AAY4s86JBMDqGRVM01TKhTp5piAsAFvYiXP9xU0z9aUc3zU/rfVQlkVR6ysHEbUPgp08DF6wUswh/g9Q4JKlrdiM71GdU8NxXxPVmc3qqviD0IEPGs/DLfKm21VVV1+Wa+rBvDbUlsk2JJYtY/eNAr+wmrkuAy+yyxU/5QAHgSQg2F2GBSQ2FesdkTpTIShOBkaW42zJOuriF9IQZaJQIIAEukoE84AJDK696YlYj/Q22N1E2Mq62vdtW5sJOiSs+Hk371ky3Z/FgLpAavPeqV94wYyX3Up+beWws9hUbHVd/ufJQF2q3vW/cRADg0ANja2n5fSZqRbmczH+bHc2DyLKDoG2b9oLgPuDxYqanO46oGk96dOXFvDQp9QPAgEjUUF8CoZjqfEpWRIASabuTFuRGkIjjVSY/wMImURp2w9usxeQ4DAPx/TH5iADefDIh23+bnGgAsqTj4APjwVlDAaoAXYFoAuRFMqUQtAOaLK+SnVBWnjH/0AVskDuAnAIiUDsSkZ/2z4Ts7jQ1Ptot9QoGe3QI9F653qmd/uqVu688gxY37e2xf8eQgNZRqNiPAd89WKiOAAKJhIUBIGBwcFbcB84YbdeZMczByib8XSaoMAOA05yaVnmlRN/WJsPbpttr+clet/mJT8vgAghUAQLxnv9hSeQFT7FpNn2QTAABiBFgGbUtPVUFR48EYwM/WVd34hMrNq1cXb3WprS92o6YD29qGfCOu6kcAwJQ2SCz+AYATJ5pkjZvl9729w2LZsv4J5BknvLNzaJieULUSxzO+NyzYoAtQGoHKWGQRPXSWCYgVAEMJNwHrADSk9Jd/uWmlpQdjAPtEoF319EfPJAhoRc1wAGBwarOZJ99ckc3O+1wOpMAAkoFvzYvsF1kATvmg9VDeKiQirAT+nmBitGvhwRYVNfgIADrlvvtpQ3FoePkes+E5uABWAsKs49HRsQBBrE1Of7fnVNM/qvr25oJxiZkfP7MGAb01iGqy8dlMMJfIMhAEIeVBxJqby0322rzhLPCw/WWe+g8AvGJl4ZezDjHl4VLgs+Pisl5Zt/jv5OJZz5GyWW6OwuJGIczhHi/9fEPYsMe8/JDxhazpKYIk3ExuIje3r29EbjYICwHiypV2MaXSWeFI8BMLxk+biagy989Pc05HnIV1hXtKXIr7BQjhirAOOdkx5VmfhVHIYqQDOcDSnZEo1/unLCgD4OGHTO6Um3bYQ+EkIJLNQyE4hKnFQwEk4LhjLQAMbjDzyHLg1x0BgPOHg1OZFjYJJBtqOSiSYh3hs2OBcthQjET8inWXyJqChus9hqhnXYAGuVnJnOYEUzDLeIAsFAKJ5FExzSFPEGnFsuAE4SHCL8/WlJqZaUl3wUrCAq3aj+bUTckkLjQsSlwgnhnAbcafSMMRrGP9wEQNj0ElO3ARysvbjgDg0JTFhdYQ+q89/homerM8ANNiwCQjOms+dNAecGBRYH3w4BOJzuJL37jReQQADg8sOlJnhz1veCS06+IasQYBDaxEDhfcR8x4rh3zHeuN1zvNukw13oKFUqwPuFxbLIks40/HCjqyqCBrEGNgUQBAmH64FBRYgN74npieLCgWDKai1aSD5wB4HQGA889reHhUYi7l5a3ixvBcOLVxHXle/B5XkO9Rk+I0x3LAj0+UlWdnjIi5JV0OfaVdCGykuslyQWLLSABwo4IqEcSFegkBhWgvC4kFhcloMrbMuAP/4mIAYFgCxB+wOLA8SFvl5nrvXhM89YrIBveaFBnPHguM+8fpjFWGK2fd3GNj4+LSAboAGPRgXo9b5+UUYX//aNIuxeQPVtS9NoMEBc2dwraMAwAeHgEXO6iP7i3cOnmoLEoWIZufHDAnEgsUQCOlCUCY7gaLmdoJrAoAAwoxqU9ABisC9iQbgXgG98Sp4BGxEADNzvgNmxhuBxYSfjbWEkCDCcxJjRnMvQEsCayxkcntm0BKzAYLjHtDeg0Q4PTmvbjPZIT8RrgK5uM1WPGck/lbiGyUv5tgCZW9Uq+d4qRjFMe8eCJ1pBzgSY9YRUNc4g9sZExBFq+xMTosG2MfNCBC4a8a4DEifisFWQCIkWseldMQMOF1AApuCnEM7h+Djc2CizX4OzZgpN8xF/O9GMyLzwifG/MB/AwlKcP0Zm5sZOIszIm5cJJjQRFj2Qe6FrGwCMolEm8xUoHtvlsnAFl9fXLru3d7TnWuzagz2lUV2TYNAJDa1n+1HVLa7msA4MF6R648/sGGdkOuihOWE9B6ymIpcDqWlRmgQiyDAbjg48cabEw2dKTfsUnN92KwYfkMNq3VOikoaHCcyhopFZj+npbJZR8A82TuF1qXSNjBfEWz8VKVcf03tHsUKpDrYwDgFI1VIOTVwWaLVLjk9WHWmPvxJPWjpcjA7QGwkz4ENIhsf70bzARAeV/8YN3/AMCJgvnoqL+ub1r91ISgKNz/czaJS/itTt2v+gXmIJrvt6pAc2C5pEpkQuymd3dOqmIn3lxWTTNT/gcAN1RUn4yOqdHvLkmZb7X2UUmr5NpAH+YUjVS5eAQAzung+a0qMJjOO5+6ahSCteheUOAjhXBJBYk9dmMIjjgtpz3+vWVVWb2fP8V0Onsp9c/E//djVPr2/R418dKCND6l1Nl/VYH+cxeTVQnKaCYgQRHSf04/UNRQmmYnDTPyYpuU+BYeUp7J7w+THCcqn44yz5ROET3f9V9tqalXFg0NBf016sd+8qVJl/rRCohHJSirAACfzo1uxaj/ogiMOgqbvypGhVbJqWY19daKNBtBTQgRkmg8gNHRcV9xF6QQ6FGvevqDFSHZiBujT6WWhSnfzN+vbpdZN5L+IiwP3RAouG4V0kgdtzbXCw6xNlAJbl2cltcTaQUECiMUDxG8JJfuu0j63W61/LMNVRfISyOSgvqxX+bv18CrHbTgjAMAWGHwur00J1SAKi2ghBpQpKwB+XHm77dFiFsz9cayWvt4U64VjcQSH2nuc/r7MYDpnZS354Ii3uJwE2Ud/ZMlOf3pIyCaghGkw2G2+a2m3hyXKtpU98SEaCTm+0jKzCRf+akHw4GMlAavq1c7jgDAjrSIU5yB5llDR5ATMlozEJOv7sdFCHMwWWqqF+pGyAT4FQCwYNJLZvLIjYAU4UdaZ9BS0A8x/RHd7AMAIcT4rFdgaP1IY5oL3zxyI+KR//LygJFGFuMIANwfFBz5qVfggbR013Aa144HboDZ/TedYp52EFIKC/15ChF4bWgY8O29B7z8an0xqJB0U/zGcwBAdVlbm38DOZT3+k0KPBQA/CdkejD+0unj+5/ODFKGFEakc1CO6waB6QgAopCZrvlP0ixSTUN61KIyoDTSCwgOrdOv84fD4Of5EzuyW0DW7YEeQ3r2gAfMZxRk7G6E6Dal02+ddTIJACDSIN7qZwDAAk6PFeyBxedktxTKfPMcrrgiiIOkl59PUD+yGK1FZGgJuq1IZG8crNUVNSnPAQD+vxNc7nztV/Xvzqu9b/bU8989FzXVUoeqxvwehfY7ADDg1NvVvCM9TFiza1BDdgEA0X8nfJ+Wualg92BzTH5/xd5ijvOt0lm4p3/U13noTAAAhEr9zCORNastYfd7SqTV/28I+P/2vzcqP+EAsPe7Pdvcgfs9w6LECqjsfLmjLt/p9rH52SIbyCkqdXXHsLRaR2+g0KHCF7IAfhSStQ5ETquru7MHAJzslrr4s40DALCr3QE7go0UA21/uatOBNo+zz6fU2OvLfva/3Qiik6lIbqL1meA4EiqIFB8skkAhf72rQvTUtIND8D9zWN/YZP7GocZmv9/NDB2AAD6X5q3h799okmEROBvU9M9NDupZt55dgQA4VWG2ioKfwaM+z2p1cBTlNU8PyVtrnu25lT35qzEYPyqEGwNZrpf15CpuU99M1v0Itn97XO1+smW/nra1lJXVFiRFqsfGlXPPthQtWPj3mEmFjaok9qsz4nTteIZOGGJoVYbCQB4Fqm4FDTDMEVcicNQok0Mxq8KwdZBOpbajIwHgP38v1NoWq9m3n6mbjuU36YVU8PUpJr8kyXVv/LUMzwGJMvWP9uWnnFLP98QzcN0AQAaCmRgwgHgcYq0aa6tItAQ4462XKZ+sCrimnQn8jsAUFJOd6iMBwCn855sBE7pYw7nhr3EAeB0RLLM3PQP+0ek70G6AIDRNDMZsvmfaleJOADdbQqSFH+lIy6dcemQu/zRRlCjAQDwa0GWObBkaLeW8QBA7t+p+n80+zALy644HxXG73RaxjzeUXqmRTZXkGR1ozOu2ISTLc1OvdCq57StqvT7n680nsfxk83iuyM/RofbZAOxmP/Wfg5+Lsm2xgEAMve6G6cx5+lU3rZxelKixG7VMXhm0enFM//euoh6XrzdpUZfW1KNM5NpBQAasDyIEvS7qk3dlU82Vf/z+RS623oTjFMZFGa5RyxLC+up3jHW0+kLbZKf55RxSwfAS40pUDqmXdT0j1YlRhFPtxinAOD6kz6xQGIFI+lLgOoyEu03a1OL4vu9LNhaW+KevmSa8p1O6f8NvrygHg26U5uPgImfdQDMQWMNu/UY8O8hY9F+LZ7XX7zTJbqLtLzGlTHBgSDfLX0i5sfh2/tZmDW0vLzJxfLyNFwg6idOkDaQ72YRJdcjLYnTVnoBjGQAALRIStZWN0y7HgT6Eq3f4O/IYtQMjEn6tm93Xg1+a0GIXYcFDTOB0mxmsNyLA6ThAp3Q/ycYNP/+mrrqojBEpiw4uwGAPotE6ItKk3ONiOqvfrIpAHAiwBMZ+vbioSldvzZnSW+BWdrYTvae0gSaRl5dcvVa0HP3sxKNOQhi2uaS6ec7+eaKupeiS/FkdFzy+xufb0sNASzOu61Dh1bU+bE9W6RBatkdjYk0LDa7/RuouaS/zlx0tyQXGjMFHEcAYHHvmgxiTqr8C5iMPNOe7Vn5d+e3u3F1LqbTTnFxg++fCXwAd+IAaWE62Yts7ctP09LQkq4ufm1MGb7Y7KDRUuSD2X7OplQc5j8U6yejYxITwBogRhArxuPXFu3RM0wNmQUA+DaXL9u3ac7pDfjsl1uOlZnGGiw2P2sZ2g0ArUvTjgIxVgHchrl316JmFzIFlN3jA6QB1YqL7dus499bVtWd6SkC6eujsWPTEQBA7b7SIRz9AhfM75t1/eIWwB8IL/DCLfNrt+D01AW4SVLRvjqbxjbCRG2fFPykI+iTCTp0dgEA95+6/1suSouXnGpWAy8tSMFThYVSfuVKu3r0qDcjAIDn0un44eZy1NyuVs4o+yx9uBFXYMiJcfx4o+rvH82IhQbxpLMz+YATij9jry+lZe50NCbuAOeAGhAnC5syJWOWNgB4/LjPNtkmAkNwyNOp3mI3eSZ9AJD8SYM6DzReKNjpmj/xHwCAArC7TQO2WpnpHgZnpi0zAACSBgo6dvDd1z/dVifOpi8AV1HRpmpr+7IeACi6apie8MR1XLnfK/TjtffW1PFTTQcqJZvnplTf3pyqwkLI8UscoFNKzn0PAKQzaINsx3v1bs+p+omJtD8Yv2vQBaPrJ5PjniPJhQ+e76HW3NQMzGt3ZOOzbXFNxJTONaokG59OSpxi9p21tAWOk6nTcKpuxlUAIJ1hR/cZikZA+XQvOrgM1693ZC0AQL2e0yftNQ8G3IgBVGnrbOH9dTX86qK6fLdHLX6wvh+LqumVMmU/PBsnK2ddBQA7WiCLzNePnZP5SjRHe/FiW4YAAKyzxE6Zh32jauiVBU9eD7GmK1c6JFBcNzEu1sDOV3vBEnGqRXt35nzzfOCbONfzwKWLwIxJlaEFFxyeuRf8N3zmU6eaMgIASM8mIkOFPw356oRHGXfwAKxcAMhiG59uq71vnqvVjzclaNn5bEasgPrJ8bSQyBKlnDvHbXCxbj4vL/l0BrXhPDg3ZL7iGZRrFhTUZyUAUKJLFsar18PpjxUQJA7V9kuqcOL7y9IbQtrFWXQK0RXM97CWoLP9M124AHKzqSrONM1MuSbzlS1CIMkAAIKc+NZuaS4ku96sG2b4O4vCHhT3QAN3JKnyGx7O6KCgzXpzRnnahQswWh4ln8o4Xd7qqsxXPD6zu8qtzo7S0sa4hE0Q4iSYdtnj2Q+jXfj+9SAoYpYn146PR25W0u1t7gBMzbOONLf1QcCMYJNbMl/ZJASyDwCRhTTg2aPhP/TKonTioRIP898PDDorTRu2KLwRSouXPtqICAAveDyjc/9+j0N6h674y2NJlzVeedAjclBeMjmhND961JfRACBpvnfXQpur/n5Pnb/sj9RneKEWpcUEkZGNa1t8GtIwttEjRKbYxLN2OUh9BwCYy1ZzLNFc88JP112V+Uomyuz3EW4yC28jSl+/Rz6pfzhMF6BUm9OX9DMs8UlbdypoOUjtL3zzcAEQuWa3Zb7iGVSbXbnSntEAcKt+ICIAcHr64Zq8pguAalXNwKhqW55W97Q/n0zqkWdkB5U+ZLCYYek5ldLi/ZMpAKLIhMDfGQ+SbfD/nSNmeAMATl1o1ebxQQC4Wb9vhnJ6YlZDzMov8lYazcideyNYSYpx6YMNNfvjNbX5xY40N0XdiM5NiRKcOFBtKaDScxKAhM6KSg9CHZSEVlW1ajMjV78oJzBOBf7osR5nAj87Fhh5elw9pAAoOdTqWJ1Ji8xXPIMMAKmzzAGAJg0AB4OANXpNQJ4xN78EAHP2TWjIQHRIJsoOLdhLNQFeEmwl7jD11qra+mJXLIHHw2OibEQPhESvycpvSDSVSCAegRFo37gTBjXfQrNFgOD48TN6QZSokZERfdI1qby8AilHzM0t029SqGpra+V3/f39+mS/FgCEu1FRBh5zoqIZZdpioPmjVxla5GSd42anR9sgWgntxJsrqml2UvrwWX9OOs0K0OgBXH/incAoFlpzszfKtelkTDGS9G3MMVrX9WswnXprJeGajXjjaQTd2fBQ8GGtQlxjw5NJQAZ+f0+G/OEtOdUbGhpUa2urAEBOTo7+w375t6amRrW0tOiNXajOnTunN/ew/Hvs2OnI1WIaOBJOlyEr/YMVz1ZrAWqZ0IY6HgAAiKmxz4sAdvQftPYdpOjG6h54gdzklaYtEHiQrsMCWP54Qzo408CG7s2J1sJEy6ihhIzWJlaPccKPyt6j/gb+QPRDOPjFEz0KNGpcFAC4ffu2AIBxwt8XAODkZ8MXFh4PAgIjNzdfPiSc6osfxkjkIllE6ZL5ird23lSTk0kAACAASURBVL22TW5FmBsiqhv1782rR1HAjtgMuXX4AV3rs9LpF7q2V64pHWxNNnpldY+oFIW7Q/zudvOAmP20bKdbcjI1LZziaFGwDmmFhvuOFDouPPoU9BU8nVBmI/hFub5p+dq/7RYXIBQAbqrS0lI1rs2+/Px8/f1tPYr0BK6rtrY2AQPMDJoyUPRD2286/5CKgcd8wB/RNwcFWQIj428sq/KbXcFgCd1i0yXzFS8vu7FxIMMAoFEWUQj7kiarn22LzFZUMNRmNgQtTjI7uvvaPdyR1d5nST59e1XMetwhANGue4KFRqAecx5LDZIT+4ySdAJ5qcWj5H/3xfSvrq5WVVVVEuALBYBrGlVOCwAYwb+HepToD6/Up2FX4Gd14hfDw0bNlFbNY2PjsrAgMAAK+CQsNko04WdzitBIgmASpZv4lemU+fJbcMlJAGhffSoCGn6+LuizbrVuv/GkX42/vhz8nuBooia+6bsTv4A+39AwoAYGRgXIOHTIahC0t1eKTv53SZ08eVJO/9zc3DAAYHPnh1kAd/QoDloABkgcNJWJlnNBWAO4Avgk+CZrH6yr3ulJAQWQbUl/X6nRjBMnnTJf8aaXMqEbUPiiY6FZ03v4qcd9Xu7MIeROf706kRnDZQqmimcn5UCLZ7OzsTkkBwZGZPA1RDPmXhIWCLejsjYCAFyUiD4+/sDAgAwCfKOjo9rk6BMAsMYA2PxsejMGYKQDQ9+Y3H80eeYe7TMO7s6pu/e6VJe+mN2v9tTae+tq+tsLYuZg1gAgXpTcJg2TKY0nogEAUepEO/t6cbCW6LHnSiD1lCGOStUqG5/4CE1SDf/fkPjGeoTTz0HI/Sb1ymZHWg6qb7w9JuzQ1ggDgCvi/xcVFQXH3bt3JeLP1wQHzQ2PVVBQUKDOnj0bMwsAVz4aAQjfiKjo9te7QoioG5tQqx9vqYv6hhG1hMWF+UYKESuC7/k5LgSpEG5ouhZVc/OAKitrzVgAIJC38dmOOpUB10hAjA3nWpVoWYtqX5xW/Vuzqq5tSAJ0pO3w2Qkcs45xj3GTC1LgTHBNqaprhQEAGYAHelTrQQygVH/AnYALwOnOKX8+hAeAlRCLB8DGPSwaSb/33Pw6NfNOZJkvLAACHGx8zG5uKO9LsJEUDz4SJjnpD3y9QhdEHfjcTCIBGQDQKKanSf7xehwm3sG6qXcgNQmphhMYSxBTnc9gXXBgsdH53lyX1MLYbcniGthXGBTxh5clym+MxJmA3KB4CUD32hOX+SKlgovAjYDYUFvbH7QYiDEgBvHwYa+gLYuA19pF3DFIQPUZBQAAJ4EmItmrv9hU5zOkziGVjkec0Pw9pjmnLYFf1hVReO4VZrhxonfKGuNQcMtlJWhrArYni4EIbMQjYQTTjwyAnTJfLGaYTrgfmEogJYKXWA2YuW1txoNDoAQTkbQeDy8etRUWBSCTSZvfCgCIZqSrw49Trg3XFY0ngDuJOY6EGCc2BwkbG7LN8PCopLY5zfHRWU+s6+Me4ToQPyi1Jc3oSLlsd1wKQDRqgETiJuHFNN1wKbASCMhQC4+fRirMfOimijEnAO4FqIsIaKaRgKwAADutsrrH99cDmAPqbG6eK5ageRjATeEg4Oe4k6TXCOyyZjHZIbTZyR0wqwDrpyZEnNR71Y4OPABu6mEKQGfK2zwl84X5BrqbZh/RY5CfVBJmJItmbGxMuA2ABMCBWWhaErCzOCEAC97Hq0zGyJZNvZremhUGppdVfgAqng9WG5sVgAaoyTZhcQLOWGg8I4JvbHa+N4NvpjtoZ3fqeKoAuzdmhVNBZsAu5SHWHAeYJwEgnhbgcMe9JPOVSDUWi5BMgLkIqbCChgkoABZYEnAmWHwsRH5G9gAwMUGDv8OsJI5RVtYi70nOFzejwOWqOj5v85NN0QBwbgPXy+ewiTFdAUoAE+DkPnA/uC+c1GxYNjRAi89tbmqzYpX7jLkOQAPUaDNwD9nc4feOgiB+l64qwMnv7xf8UAUIA9aO9+b+2dM52IHa8v5DGFBX9YbxmsxXvK5NIkpALEbuhwEYLbLZARFzsQMGgALgAEgAFgQxGQAI//IzE0QMIBmUDcDfATq8BycB78eAXEVgyhx83mGjTp+c259tq2vXD/6OzWm+lzlv6zCvwRzmtZjXg6/KBjavB//avB5jI/fJ/M15s6EB1n1QbEop6Ip1YFcNfTJVgLQhM61B+BXNs1O2WUTEKVIPSNuuXdYmCyEWZxqziIIJv/mWLHY3FxOmOQBigogJJJycgAmuymGblDkfNhbfeaYWvr0Y8XexwMUEGOZiDiwjc67m3AvSqBOQTvk2swpw6gerqv+leSEKnbJRSAYXhziHpwDgMPVSAiJelPmKVwkoUnGTnwcZGBbmyOh4xgU3TbctnQKugAA6CVCFi23mj+ACpU4IsvmCKVSIhkpwzGGZnfFpTz2CTPiZmbRB6JFXNzaekelNMyWdSRLu9hOCbI7U4pdEK1ToXPOuzFc8Ax+2sDBzlIAwRwHkkpPN8twycZNkWhOX8LR2KoQgQ+jFZuZVtDw56Q8kkbzeiDGWwMRoBpjJSKz3PZ9TffrkH3x5QYJSmdbqLJyVmqnXxjD6HzQmfQDYCgAEoyLWymvLgEDIi53+ReJIyrl+G48Gxg6o/D7U15TJAGBabn5t5IogC52Ml36+oUb/ZOkAa5aAO1mTpFwICY7aucACFYCYFrD8ID5A9hl+eVHNvvtM5eT6dxGFN5z049j4bPsAACx/uJHxAGC0cvdn7KZ1YVp0BigtflFb13QytpLMTJZjMu99u2nAXgDgRuMGtK88PbDQMDn9vIhA2SdP/NsOjEWzG9YWm4FOPeXVoxmaBWAQKEOU5sA98cGBBDvTlMxjUDtzwqIFALWd+oWktC2GxuwDAKrtCCSR9mBRhS80rAE/LyKQttrjXXEjmpDHG6X1NQHYnW92DzwX8tQ8M7jxmQoAcBms2hRInKM8TV9AStHPXfZu9WPv9pyIjOBGX7zTJUHbXEuQ3bTekmkdbrASbZooSGRqlW19dXCh7f5mTxRmafZ54lyL7xaRm+oyKeWd9YKgfTeS3Zwem7/ekfr+6o5hdfV+r/RbMJ8Jpb8If2Q6AFD4BYnJ/J6mHKj24KpyX2hq4tnY05lmNf3WqtrWe4ogOvvnoOU9nJT2IQ1dbAMANgeceL4eeL5wAADojNK6OC3lplwIC5PAINJTD3pGVIV+QMUeFtqAsgrzzotzKz3TIguZjY6uIhsfAAAI8sKCXyz6C9c7RfOvJOAXk76Fa5+pAEDVnLWjDuuPzkYmYOIa5eZ5O0hYUNwQtcAMC+d6EkVGT3/0zD4AsJYnYnaSatr5ek+PXTFjUP8Jj27SeKKqeVAWK3rpMNJwH1jAPVtz4qNAGfZCB9fDus2m65Sff38t5JQP7+ATbax8tBl8baYDAPRkawAXReqG6QkBwAd9I2rmx898fX3oGSQTn2K/2QYA1NTbIZVVVNKkLsJn14tZLIbXliV+gHYgfptYDPqh8RpQ0U0BhpI0chiinfLch2T8P4DjbEVbVgBAaWmjrE/z+xPnW9TIdxflPk68saLOVfhbAQlhk0TJTqyZnd/s2gMAplKOkzXw+EKcemx+QAAwwC8CxQAJwIINwobIT5KtB0WZz4gkiImPfPpCi6Ri+BynrZIDp/znximPxXSYXPfrr3+q/tW/+rcyVlffifjzz//st+IK8PPvfe8z9Yc//CHm682fFxW1qMuX/cWHyPQ0J/uO/ZdI0RUguPLxpj0AEG5iuXbhuXWyWXETcBdwGzgZAYalDzfErWADsWk47WKlfXq0m2KNWTTNTYUo5kxvzYR0ygU9rekZW04q7ZcCLsybPnJcC01UcJXiSVmdPNmh8vMbtSs2ourqnskoLx+w+ML7P3/2/rdURaBtGz9/7bXXYr7e/Hlj47r6i7/43323SdAScIvGjfLP9I9WBbTJsrgR9Gb/JdKynu5bxOBsAQBTmcVLkXA2/E1q1CeMbkOcomxa/uV7fs7veR1NScKDlgyTdQWJZPvzg5mNVH1H9BCspzwpHkCMeRUn0VKdjckGjee1+MFXLazN8fH4XIDDAICgI2krFn731qxYbt4gA7lXyDX7k2eir4g7Wz8xoQa/veD4Z1KunUjDGnpwGo1MbPjwVOiIrhZP6E1FtoGsA24EfdwIOm59uRMRALrWZuREbhqbiPh7gm+JzoHOR9ZTHlcG6yXeU94uAODhWxWA4gWAR4+W1T/6R/9zzNwy18YJQ3cc+uV54dkjU+dGKTdZFrIKpjuMqU2HZXcqA+OveqzpH5XGrrYAAAUJpaX+1cpHCz/SBp940wg6jnxrXu1+fZBFt/2VkcqksUleFK65ecoTo8AtYTHwnnLK25z2/Gf/7C9UbW18G47edXctPebiBYDDxsJP1/eDajl1AnJFJelfG+FkICcH9wALADDAAnBD/wIdyv4EehGy+WukI3SqTDPtVw0N+buUlBQlNQvWzc1mNXPDmFatY2NqLwKV1hyi/RbQhcfnI1hJcBI/kH+Dp7xHxELblgz+hfk9WQA7dO3hedwLUFNh2GEleYFyyzNMtFV90nUjlR3iBsA0pOeFW8Q3hGuL48yMYQHelKYpNgguoAPn90gqZcoNT8f1ZoUlNhGSSTBPD1o/kQWY+uFKRBAgkAhwrH6yf8oXlXqz/Jm4w2OLCEi8AHBYFoC4CQzD+ffWJaZx2yOt1MmVR+tVmUk1DxfjFNvBci0X5S4bkLU6A7TkYw36BFhv7MBLC5EBYGtObmq6TjwzCxDPa0c2NtW//rf/Rs3MfF/l5jbEDQANDWvqL//y3xyaY8b3zXdACxBdiaFXFiTK/mR0LO57bRxUgxm9ThGtjdfKIQVoWCY2bI6KivaMvrFUW1m51tUdQwc2PxmGdPu6bEw2aDyv/Qd//j8p879/9+/+D/XFF19o87HFFgBw0lWD93FHW5wXbnSK23W/ZzhOcMxcZSBzEOSMR/4M0GS9GgQyGxhypaWNGX1jCXJaW0Lhx/fuzqm93xsxAf590JP+xRUvAJDT/9u/+zsV/t+///f/j3rjjc/ViRPRg2U1NUvqH//jf5GW68O6MvjrgVSWdrFQNYpXGShTdQ9D43GHXyOp2f3MREoMwHrHGYBeYFnBAoxkHpNWhPTRoP1puxo+pDL++T//S/Xkycqhr/voo/9Bxfrvv/23v1OdnXueLHoiroB7QcC1Y3VGtczHf98T0dEnjYkSD8FSP4nYxiMRBgMUFyplAIi3CaifB33iDhNeNNWO/aB3eOpUp/ov/+WPwc2+u/uBevvt36g//vGPIS5BXp43xU8x+eFuUNYMj6M4gfQzLkA89SovXOuUrBCf9WRkXKoHvUJossMlv/GkXw18az51APAaA9CJgcJRZxxahgQAawa9nw79wQ++CG70/+s//pVsdCL7f/M3fxP8+c7Oz5LOArhi6hY3CqHqWIJpS7JV8bQJo5Fn/eRESMrsjk8yXUZjmNiBwAf0TVycTh0A/MIAdCOwUna1Q0psk6nMcysLwOb9q7/6T8GN/vX/+Pfk5y+99GHwZ//hP/xH7Us2ezIImOpAEyCejrr0rOzenA2SmUiZXfNJCjEeqxyq9qMgdyelAOCILSXA3q61bo87fzzx5rIEprwaBHz+fH+j//X//Z/U0k83xQL4wx/+z+DP33zz157NAqQ6UAWKp5MObgVsPp4nkmEim+aTPpbxxOVoBrPPz0jyg4yo6qgt7DGvy0nFm1ulBZRRYeU9AMAysG70//pf/6j+7F/8Q/Xyyx8Hf/af//P/JzGCmLTpNGYBUh3Xr6NaFZ/LCrWbrAO8A78FuQ87mAG2S3e6UwMAqwZgJg8WzPXr8WkBslDQbydH7bUsQHX1XIj5H+m/Dz/87zL6WULmqq/vz/g1e5hrDlt1X/MiyQ+5oRc5FNlMv5nhLMBDi076R8TE8mQ1ZHGLeu21TyXfH/7f3/7t36pLlwYy+llC5kpWQttPI1JwnpR15+qMuDbUKFRU9aQGAARU0tV33dWiGW3lnD0bfzEHMmVITZ04613lY4J8Gxvvqb/+f/86CAB//uf/JK4cuR8VgfYBMLGKOb+O8EAggWk2vpW5SmFbmezfJD+EHoCnTzdn/M00tAATC3TSALVxetLTtQCMvd88V9NP31D/8l/+a7W6uhYXAPg5CEi8KhqpK/MYgaMhGapItSttK0+TAwAQxWhGUJfxAJDMdaK2C5Ekv8hdMk0itQAMhFbNakWjf15DRgOA5PS1BYAlkOnr1soIpAdgJADofT6XHAAY5JjM96XYEMlqHdB0we1mqIkCAGXLpqsSLwD4OQvA6OgYSqqJhh9jVyYjEJn+nQjNeoZeTbIxCL6/tdFCpg5SKclWkJFmWfzZhqsppEQUgUS55mfrQZ47RSR+7aCbWM38QEJB3WQHegj7ylG7IgxiRwsytBagKkfSs6gbn1D9L82LGM3d6u6QbkiVL/YEm8MSBOxanw0Eq5NKjfX6ok1WysouKXYEpnLtygPvUqURNS0LyGS5qZqbzmF00el0BQBQo8ZdpkychrmoBDkFAIi8IvMGEQ0Fqp61mYMM1hxDkzJUvi6JSdAlh82R6Ysl1Y7A8MdHv7vk2evjVDIJIfECgJ+zAIkSu1IFgDtNgyEVeLu/3a+wPHWhNSgZR4T+poWfEO13gAjS9GgiQDuvCgicnNR7cVef6hu/2pafN85MqOWfb8ozRZOQjsJIs2GRXjgAfkmUx8IALCjI/NMCKyeVYic0BfGzz1W6Uy+Bjn9//8syHj9e3o8Cl/VF/Pnij3bV8u6b8vOXX/5OEACivZ6fv/LKL3zZF2Cf2u2ONJgVABAyofgGBp4pyLHw/roUHEExpusvytTnr7TH/F1EC4BahbdW1M5v9tR5vc5MFWLcu7GFKbX68VZQYh5JO7MnYtIAQGSRCGOmb36DPdetbt1KzVysHRsXfUA35vv113+m/uk//QsZH3743wd/jlR4pJ9/5+O31f/yv/5v8vO///f/gT7dG2O+3vw5n+Noui7XsJ7QLay0udqUHHk8xV12xgAYlIuXB9YS9GIyMNbCsc71GdU8NxXzd5EAAAFSyqNRIaaFHlbd9he7Ynl2DY+pTf25qFJHr2VIgk6ZiP64n4e14WnS5BONvjx8OyTAMe8i+X/JDhp3VAViHCjKFjmYtkQfkE1NZuQwkhSVeMh9EdSCtooQq316CO5ksKwWAPLgN2r7xAxH1OT6o74DLclrR8dV3+58zN9FAgCqFJH34sSHhr7yyabo/ZGFIgjYOWP0ZgAkaNJ7sOFMgheG/+SWvHK6x7o232v7Q9OAtxr65WYn8j6dazMiYOk1AMCnfDGQ5XASANgASx9siIgH18BGiKayA3eCqLmpykz7NTuCZ3QCvlnbrypud4lv7HYMgIEvzvpJxQJgg1vXACrM3NdI2gikAc3aBw4gGrbs6wAkCQBupVG8MDb/dEvVdI+kDAC0H1vTflkkMwwEn/7hqrwnJ14skQu7AaB5dkpq350GgPv6BN/VPuptvXawAra/3JVFHQ0sAABT6aeyuidEBzCpIiC96aFnswGWP9Qn5A9XHNdtsGYBADW6MJF+w2QXPz/wvIkTIT9GAxUYe7F+J5mbt5/JGrTG5KbfWpUuTMQa+Puzl9ol2Feh/3Z0YUrmwNrDijC6AaUAAOluk+3m2NKIXa19Kx4gLcU4tcIBgE690rI7EIF9aOGag8yPR8bkBGNRcyIgLcXrkZkiyrvz9a6qvN8jC+P5756HNOtoXZhWDVOTEQEg1uceHvkNVI1NjMtwGgC6tEm/+slmSHqUGvtor8f35x7jCnAPMZ9T+XxkxKtaB4NqQrvaZD5zvtXRGIY1BgD4Edgz5yBkuvJWNf76sjwjrvW2xa2O9TvkvAgss3budQwF1wLmPfeKn3OgEDuhKnVLH2IACPEBOAIHlasTZMYZGgCZTxiBL76lb+gjvTHwqUBZNhWNHoMAEIjAYp6RWyUCy4a7GqiS5IEMv7oopxo/I++OP4bpTdEQ6Z2pt/aJOwAAD/JQAIjxueT1eV3syO++8g1WgNMUWToFcepzbcyZdNjDvtFDiVQEtrCeUv388TeWg3wM1u629ocvO8BjwcUgRceGRyIuzwPEKlL2ZWWxwO5IAyBq5RguAIuV0xsTEt+MTUpgyhqBtTan4AQ35aQAgGAUW29acrNbeiMACOIaaFMNUKFBA4tlT58W9eMThwJArM/ld1gFsSO/xsD/b195GgCA+NtKJZwO1fMgKr2hTyfmPfKdJVebp9xtHZJnBkMOMczl99dsd2MRhmWNYG1hZeBu2BH3SV3PoldK920BABhU2aABYNY7bGi/fe7dNVXdNSSnKAOEX/5480AE1hz4mERgTQCw0j/bV5+qrV/vhj6g3hE5HXe+3hMAOGnpIxcNAA77XJp+xo78GoMMAJkAs3jE2vvAkXr88jZ1+kJ64ke0Qm9ffio9GmuewAa0t5S94m63dHoO1uQ3DYiLlu51fDhtP4E3q6lxh0aZzkGahsDTZb1A1qXHnyGigP+G+cqJbboAsSKwQQCwSDQT0d7TFkR4BJxgD1FvNuwJC8OyZ3s2IgAc9rlBKyZq5HffZKXNmVsA4CU2oJUnb4vFqEEWC4BYERbdyHcXgwHWdIugxFa1TlAc49y5zKUAE6QjOAeZYlOb+xv6a34GmlM4QZ0/WvEmAMSKwEYCADYsADL7zjN5Pa4ATC+rVlt1h/GwTr3QKkGgSAAQ63PxmSGcxI78Blhx2i9moWYbACQi9JqQhfykT81r94LnDvnLdPWSqkQtsudZUMpOSXt0DYSEKMBjcXdW8d3QNwgEP3PJOJ2bBka1/797IJcbKQsQKQIbEQBggV3tVDu/3VPrAMxnOyHpME722XfWJGsw9O1FAZ1YWYBokV/eI3bkdz/QBtglK3zi14HUe2OjN8ls99qGBfjJHvBsrz1OHagQ7zl1qik1ACgtbcpoCjCnJX642dG2HmrlL7ecEWvYNUo2033N5JbJTGQbABDf6ejwnp4FJCCCzFYKMUHoEykW3sVmtCZEAc5sRVVOSwI3nPpL766pfrM5hN0PWp/iBO1y0qyoRFDOtGaQks4WfofR7s17XZxqRyciKvek2pWIepboRW0eK6NMqwJQcYMU7/TtzanxnVl1wUHGI3n862kWVSGrAakk2wDAq9qAD7VVGAkAIP+kWgDVHBVE4kWn2j5VUZEdFGDTbzp50jmTGBLQxBsr6T0JtckPtyHbAMBkPnpNG5DAL+ld6+aHg2LqNqbCaYle/xBv5Vi3sxvCa8NphRyi9DAM7eT2JxP3gLUmKceeYYnzZMvzJQbgRVXrCzc7hU9AtojALuW88UrMs6YapifU4gdGWzMrBZwO15EBL65qrjq19L1l6SkOo+tChsuB7ZuIzmY8aNDYnUScAd4Ap4Ud/eoIMvF+2QYAZAHIBnh9nhQFoQ4UTydkeAfQnhGgkZT2L7fFrZXCr+ZBcQUO/B3MPqSvYplD9XozrH24ITRTcuKw4QozWFrZrSARQo6kAksSOInMLsTUhfOAK5IktBAA7FidUbvf7ElRSY92AUpLs8cFePSoT/gAns9O6VMdS+B+9+FZI4qerPECMjzmYU0Q0CpuA9hLZgDZK3qJVVZeU3l5eZgElsH3l9TU68uqZ5aqsTN65GhLYFpdvH0x8PurGbc43BKNEHLV8rRU7yUSPITbLoSW+71q+aONhD8TIGfjW33Nhe+vZBUAENC+c6fLH+tRn9wQ1ADtw8q70Z4guwTlmvjO8ZPNel/Xq4e1/apbH2rE8kjnU/xlaAUE3+CMevDgof5lnzZ/R+Xfu3fvCiiMvjShBnaHVH5BoWrualY7X+2o4clhde3aNQGEY8fuZtTicEs2SnLS+kFxkufHqbFoFg8FffiACZ9QfONbCwejzb97rk6fbc4aAKAWgEIZv8yXwi20AGI9azQUCCxvf72rtvUe7V96KgfZs3fX1K7+GWNcu/DHQ4K9FgA4d65MlZSUqIKCAnX27FltBg9o9+CyunmnSp88U2rr11tq5+sd9aj/kX7tOTU8PCz/Hjt2OqMWB6ZRKmrAiQ74+pS+xqXHsD0nBTzQfaEmUzeeCNuR1N/iBxsR002XrndmDQBcutQmlq+fmKpIf9eOj4dWIerNDEenurpHJOxh6/YPjsq1Yd0/7BhWT3+0qgqLGtXo2Li0r78Zct3BL+4GTPxcMf8LCwtVT0+PKi8v1zerUpsVOWpscky9UP6CtgqK5OSvqamRAWBQd4yfQTyBbIGf+6/dutWVkhpwUnXk+mRvnpsUPnnM+IQ206lqQ1QDHXiKlyJtcqrT8O2bZqbkxKeqkZMBawPV2PDNTzbg5KnsCQJS09La6p/SdvZTuT6YyAo09YwENzvpWw4rNjvXFJ65qtFmf+vSdDCT1/XsqQSfo2QBytW9e9VqcHBQjY2Nqdu3bwsYVFRUqrtNd9WqRo/68XpVcvpF/fMibUZdV21tbQIGmM0EGSg95IMoQCDVwvfUI5eVtfim8QRomqoacLyDFA8lwiFNG5djqwjjLiD8AZJT3krhCZFiat7JI/MvVgXKOhQXoQ1AP3jTfCQy/PTHqyGdYkdXpjVwO+sCAExUH8KCRDCjKI0xB4JgZD68GoRmP7Gp2T9dXUOSlWJf9epnuvrRpirXp348+8msHOVg6ZudkvqQ0NRzGADk5xeq48ePi+k/NEQftTOqc6UzlJzw5a46WXZBm8qVenJdgYDhweglBQjUI5NpAG3JrRNdx79mk2FqU67oNWCgUgz9eDc+q315JqI5DjCkssnjOgUr2iVqTEOJLhd4Hij90jWZgqvOZzNCvU7XM6aojRM0rcxTDeS0njd1NjjV2SMM9suDBz3qml4D8BWsDWpxAaOVX8VrdAAAIABJREFUeEfUBNDvTXpw4Z011TY6HosH8FiP+3pUyqn+8OFD9aT+idr7/d6BBdq13htiASTCSkKiiCAM+gJYCTwIIpMAAzcC0KDzkFMKNV7KEfNgIgEAmxzkZtN0b8yqJyPjwh4kDWh3ChbiyOD0hKMAwIFA5sEEKABONA3SuAGxUvPynK9uZc1jnrOu2dSsc4qvTCuZfYCVzKkfz5qn2efqn26qqR+uqKdvP5NDJB5rKmI9D6dHdcegetjfoU8aRrseXXo8VqN7o2r6temIC3T6relgDMBIB9pj9nAjuCFtbUNiLXCT0CHAFKqq6pYYg9NWg5sdZCn3Db+3mOSlLnawhQ8wrU9jpy0AgAbde4lqdw4HO+Wka9hJf2Y9smZYn6QXWa9E4Fm/nOisYQ431jeHSyo0ZFqHEbOxrhl0CA6z/nB7uOaQn2NadjybVm2LnWpga1D/26aa55pVrz7hF19bVB1zHSJnHL5Ie9Z6XMkC5Oc3iAkEeqLiQh6TDUo7aywHvibiiUvBa7AuELZIJQhp9FZ3PiCGeMeEPuG3Lfxv7vU9l0tVsSxW311zHAAqNICjaw+/HRHT8y4TcdgguDyIlKKbwKY8GyfNFlYo6wrLNN61CCA4oZ9RPx25avDCjc5DrwFAyg8RKpH/PdKmUIk2TZrE7x8ZGdHo1SlxgJycHPViR2coceR3eqF+vafu3r+bVh5AOOriu+M7QXIgaNLTMyJVUERJq6q6JOaAKQbq58YoxeVvnTQNCcLhwyFAgi4AD+Tqg16RHTuVBnoq+WNyxKddsjoK06A8hOTa/E/XQtbw9MuLQXFQXBSIUKwP1gnrhVOc9UPwjTWBRcr64uesN9YdAOJ2DAsRmUgAcDUOvU5AKhT05H9P9DgZ2Mz7TMCcnEK9qZ4YZkdZs6pqeayu1lxT+UX5au7dOeEEvHD1oUcplHVyouFW8EBhfgEEPFCAgQc6NDQmVX/4/LgdPFQCMiMjoxLAxEyzNZ2p34tmEdB4iYLTrcYr92tdn8w3MlTwtaCgXjVORD41h6fGxURnPWAes8FZJ6wXAsHmgeFGrCAR6bFIMaPiOCw4DsmrVzviKwY6e7Y5pgz4wMvzYhlU3vefTgDmEMjNRi8vb5NoK2YdJhyLAf8NxB8bMxaIFSgw8UjRsEA4QXBRMA/zY2jAYe4T8Jv7yVqIDqBXxtzrS6r56YQvnh2bkWAZz477z3PgefD82LxE03lePDeeH+b58tvPIgJA58S4rAO/8VZa5qdDsnLxSodxn+7f74kPAECKw8QTO1dnROm2un04I04LzKP29qGwAGWjWBPEFzD7bt82iELcG/jUgAVuB/4VCw7gwGwEPFvahtTUdxbVFhLdy9NCMsIiKS9vFW4ELgzvjbWRzClDVyG0/akaO55CHn90+ama/v6KixvYuKeAJ6csIMx9IUiG+Q1Nl80M6BIQ5n5yXwFn7jP3m01OqzpMcha1aZZzX3lvAstmC7Ano+MRAaChfciX65TrQsyF+E28NHJTDzGU5h6LfywVRIefVuSpCV7Vjo35HgDYmKlIn/FgWHgn9AJ8ov379V9uq7FXFtVdvUDJYrBQWbB8BiWa+GQsZBNAxsfHg/4mJikLn9dgmvJ6/g4rRaLMkxNqQ79/n968E99eVM8+2lS39SZgEwHeiYyB4VFpU3bjZteB3/F+EKPMwXVgBTG4Hqwi5sO8zOtivsa1GZYUgbJRvQmN6zNAkt+Z18bfmRuZ9+czDaBsE4CApGRs6CTjDhpw0F8IMf9fWgg2z/RdLYMGRwrDEv07LNX+/pH4AACkiDcffrdlSCwBctZ+BgDD6kmtDgCCDUUZyZr7LHIWO74nC988JTnZ2BDw2NkcU68tqzbt25obc+XnG6pBA4a5KRMZmMurH6yrdg0+4b/j/cwNz+CENj/TBBvmw7yYH/NkvszbtG4gvaTbjyYQWK2ttfqpCTX77pqq0ffKT3TgEP7I68shPQMTGYBxQdBqODQdFn/EFv45+Ul8Xb8CAOZ9sk0jOGUkuv+pEd13ugMt0WCaj5r0YIKLZ5LUMeQkbluYCpEhT2uMxmHBVIQ22URepQPHPGC0m8Ozzk0SUEP7e8TIvyfTCPT8lQ6189WemtWnX25ug+9uLqfdzUQVjwLRfYps3Izus9kRZ6G5x9IHG6prPXnrCwC4U9svDUfSCsDa6kQXn8Il1KecShkCmIixjM9M+m6N0mLO7OqczMCq288ERFtcZ5LXTkfHnGYVq3pDFJY0OO4L9e7Oqbalp7b0nSPwdPlyu2vmfsoMypImURcuT7F4iWd9Tj83GoqkS+2JNBYiFtxTrCc473Q/cpIBuaTdKC+l+OJ53uytkhT0DEMzAQ7VxBcWN0nQhfLFJ0Nj0qwSiaL7vcO2mcY0yaQoBgINddKwzFJtqUTcI6J2WhRzHz47dGqnzX3HT5V2g/489vqSNNJMlyAmohdWcKfgybHPu96ptj/bVqWl/hFCoa06B16qgjcEamMCAH5wqpJJufl1YqKGp17qJ+3JNxNwNHvpSbXZmyvq8r2elE/CmGqxFjIPAFCcIUKaJgAQUyDHnBZmpzb3iZ9cq+mTYqGRV5ekT4OTn7nxyZaq8okwCO740s83Urb2QjMB0fjG9f229FCPBACYMHbcELr1mos1v6hBxDLRu0vlPWM1yTTNfcQ4XrjekREbPxwA4JPjzqRrHgSSZ955JpYd/IZUGmzGFU3fnlMj317wR4bqQa9Y0raoUA2PBTIB0U5XqQ9P3TSCAHNAfebrvbhkjg8bJWeapcKMDQkpguCIEyWiVnOf099pufB0AgDRdxGTzBJ1oLrGAWHSFfqgKQoWUVWUtKW1caz161jP26gJiCI/xEaww68dfmXpAACMf8++NCEnBF1uz11OvbKMzMfIyFjGm/vRUkNmkQidiW+mgSBzo7ZPJLDJAOAKDL+6mLK5e9jA1X2qrTrKk5P5+/n31kO6+OCG2rEWDwTlyw3x2GgWkXXTs2YPqwuA62JkAiJF8U80iQVgx8Shp079cF9+auad1YM6dh4ZlACbHZAz2dw/DADgMBAhd1WGrWNYMhD8y/rgREYrMdF5JHpoERFvHxuXpjfJAgAbThigev3gttDePRkGaazfty5Mq8an0TMi8Zz6kXUv41UOSTV9Udqoik94+wTFBO6mIkyf9oh0Zqq5HxkAhqT4yySawCtwjaGnfVHiQtaArtUKM0UwkETjdQs/XQ+xUDgZCV6y8fhdXK8nM6XB/dnHm2rmjWXtQu7HjwhSk8Ha+nJHzb6zpi5VdcUEALrwWDMLSLVbxTsSmTeuFx2q13+1LW4tfSNOl7eKW8a6jDavaC4APSdYy3Ar0GAw24Xt1wREYcNVV3dnxcK3LrT6wTG1FbjRxVnUJiscALgX0n/gvDuW2kW93rAOo/EPiEssvL8uG5NeCPAt2ASmoAgbCXdBzOOc+F4/8K15YdIhrvnsww3t9iwEA8pQbNmIvA/iLDTliFZwYwUAUtCsHVPpKNF5M+D3E9xumpkMEeKZ1+8Ta16RAABNSb423YGTZS0izGpYu40BazeumuHMHpx4PLSF99ZU++BodgGfaWK2Wumhdapvdz5qwMl2Mpf2RwnERc1baxeME9BqJkODZqOYG6nSIuMez+vN05OS4tnvLgqLkrhDJB87VnbJGgNgwC404xaJzpvP2PxixwDE3x2sXGRDR5tXJADgd1gSCMaG95E043zHokcImw/kIGk66Nap4EreORDdN819t/sBeBkApF/g1pwnLAAAYu690NRk7ei4gJS5kayBt7heX2G8nkKl+VeXJNbAIUAgkrVAug0TnM2ESR+NX2K1AAAP/h5zmzhGovOG2CZisJ9uRSxd7tXPI9q8orkAd9uMawFYaCpjDQ4a7e+iVAtZZY7w38kNQ0IAUeg/5ndz3xrdNxVViQjj/hwBQJ1IjFPb4GYMIKIWYk58J6m5oeM9ec3XcxIuv74sAMBmgYmImW49beljcDmKSxweAxDuiz5QbjX0JzxvU8OftRkJALCSrIBhnddhaUDib7AqrXLiaBoeAADKUAfDzGDYWHShMR8WJA2/RsZNc5+ob/hNokji2rWOIwAwF9hHG0lXFyacBdCnEacUgp3w3MkCcJrC7RBf+mfrInqC335RW2pkDMoCbmr4Rkr09QvfXZLTkUCcudnWfmW8BlYi38cCADMLABntVv2A+O6Y3+HzEFfnq11p6hppHifPG6C7/eWe2vltaLMYQAVSXV6BASbh84oEAKhQ4Y4wN1wALI/m+ang51WJGxS+QfQiCJcB4yE8sMgJU312/VGfr839SNF92I/UtB8BgDHwi1/sdq9clrQfwBzkAXxnUZXfNCwyIuGU72Jes6mstfDhGynR18/qk5FNc0BnT29WiszY5LEAwNpejaCfNXZizmPHIuO9p0fLwnTQBWAzkylAJLZ1eVpO6vVf7Yjw7paeP63dCEpjOXAtcA3C5xUJAGB1ki0A/LifNGIpsqhdG0VvEQQxUGYJpyDypqQwHvaPyMS9ntI7zNyPtgnOZ1CMIyGNuZaD105pLq28sgH8NvQGidQxORG5ragqU7c6I5r0HfpghclKGpAUaIHLjXBOCdsz3BSr7onoB4POIBOVSOFoG6oH0C7+T5kHTOlY5n6kgUSV09r4nrKK6J+QGx0ASDlxKmU6FwLLb+lnawcp69/sqSp9X2J23dEHDC4LQb9oIiZU8EUCAAKEkgVIkyCpQXm30QwmiIMP0705K5HKB+FdSOyujdYPBt+NANKaNu3xtUTt93hjUmSegYERiYFk+sZnwXLqYB4zUMeJBABi4r6/FvSdM3UQ+2kYOrhJ8ZlpnMOaJjiICMcZy97AnJ62sFxpu06uPWSTIUk+PRkRAIgXpP/6bTwFCRqVBZpqEjxyuvfb2OsH++oNPJ+Py9yPVGqJoq+f25rHLSW2NiNtw+U56QVNpLq9czgiANBauibDuRFm6TsptQHtb2PpcvKbJzMnO6lKGrHCpiMIzvoaff1gnQvsQnj7MA0nf7Ai93ZQu1Ebn4YWxaFdke+JprgRyAHJKqSQIjwRCCSRbyT44JR5A7JGalmGsm0y7gelkUND2UECosbBGtTCD+0eHhPZ8wM5+Md9UoWWyfcDYdNQrfzYBwU+PQC6983ziCc7DD2sYNxm06/HWqibmBDGYd3keFpbo0cFgH16YPL1+ZhED3pGpKtt+/JTxyYOeu5FYEst/zw5DruX+8Xb7vNOTYjeHrEcquCwmFraBiICAOk40nO5+ZkbByAanoz6FfUSkQ6gnDw/XX9UqaDkFEtIt+DzQI5wWtkVEyv8AUDiwDxLtOIwUkOQTB1sZnjvmLOkqGB40jItEgAwYJJdzGCCVLLrHs5C+PprnPWbyGiYKUT/cj8FsyimgD5pbnxEQgjWEH+AyhqvUCjVUXShycYUoFhvegOUlUUGTXxfNBcz9dqRgDNosYn/7a3GfrF6UTEi6O30oecoANCO6caNzox4qPhe1LVj3lJeeeaQzAYiqOH8h2wascRQSVURN8jUaz/YLSebhhXp9QlIZ5dMukDomQABddREd8166EjCEPEGgrINAIi3ENDNz9AUKUFv2pVlPQAQBDvhAsOPOgK43/iebvq9cALwe4kdhFd30Q76zp2uIwCIkTmozOBKSZrg5OfXZy8AkAKkKaXT+vY1Q6NyGpMmEaJOi7u92QhUwr2mRwFAwNekKh8+zN5CoHgAgBhAow+76MQ7enuHE2qDl3EAQBqs10HmHhsPUsne7/eEHAFrkEYQ5KDTVSPA5ieAAytuUM/tsosWid8AgCwAJJdMvX5TFTlrAeBg33Cb2Vb6BnPibn21oy7e7JLT/+ojo8goPO9MChHLwC1ZLhb38jtralXPJbzLD4QmahsohLJWUmUbAFDOCm04U+8B8a94O2FnJABc1345ZrBjmnNLT1VN/6iopqJ7BsccV8Da5JCSXURH+vbmVM/2rNAl3ao67OwcVrc0IBEohNJM4LCyukftfL0XIsiQiPJqpm0AGIHXMzRTAhGoMistwMAXSXXFTWAQgMMCoHPuo4ExtaU304sw7yxU4butQyHlp7T+crKgiKDW4+ExsQCsHYEITsIh2PvmINV45p21rAUAagJw4zLx+p1e/54HAKdTgMQAmmenRNWEUz6S4CTmd+/Ovg4dMk01A87w86HDogDTsjAlbsj8txelMYg1RhCJ640Lk60AQKEXllsmXj8ZIIqCshYAaARyIs0iH9SfQ9zBTeCkgc134pwDAh16c8NvN1ssX9Bm/fanBzc2cwkHgIWfbqT9oeGPE0SlUYRdFN14AAAQp+ArldbUXh1kgGrS1BU57QBgpADHVa4HaIy4CJz6mJulZ5sds0YgtpjxhYvXO9XWpwdLl2/W9YfIM1N8dMUDVGniFEhIET+hVt2ONl40gonHAsRFu52BlGk0MOp80iXYdgDIRiok7ghKx5R1rny4qZ5+az5ykVBFm3oyMi6xAvTd0j1vQJFTODdQso2U9NjrS64BABqBKNtm2npAC6HFZU6KZwCAi29tzbKL11YPp/mjwTF1V5+gfrl+ei3CozDVYVGVsaNeP14AQOwiPHWbCeOUtjy7uoazEwAQAqUbULaSYIxeiP4xa7u3ZqX6DKl21F4rbahh4PrL47RwkHs7VZZZOXMahAwMjGYnAGRzQwzGlSvt/gJAbb3AUaCRxSmbyCuJAAAp0nsZpp1A/AsqfFYCAB1CLl/OXhos+d9sbQkWTItqN+hinE1AYGmiDu2FedsprZWdBUHHzF6ALVm7+E1RyCMAiA8ATui1gu7dsTQKqFboZ4YrQqcd+Bx2VJZCBispaco+AKAVWFFR5lZCnXqhVVJlZ6IEubK9EjBRAGBQxBWrP4TTg754VwPU9fs9wyJtlup7dnQMiTpQVj17TJ7h4cz1faAgkytHFYjag/sRhD9xgSoq2o4AIAEACG8X56qAR36DdMY15bfQ4oe05XRBVEaOTE9/cFKZBTzoA4rpevTgUwYASFKD305f27Dpt1aFzo3UGz0MzZbb8Q7a3fE3ME5NwllWHgQ89Pr6zGVAUXZsdmuhvJiKvvBOQZh+2VkLnjwAcC/hI6RLBJM4BE0zkeZGXCaRMuWrNb3iQtxrGxIQofqzQLvAuILXr2dZQVCma+HB9EPIQnoVvLkiBUYH1WBGRBAlmwEAGmyiLeEQUrngww0DYFB4Zn5PlSpBRWThqqqyrCAok5SAo/H+SVvhs/LQI51YQ0NjqrCw4QgAEgQAes0/9mH8iHmbrdGIJ8BsJIuQlcKwmagEnOjIlp6AdgMAUfjR1/zXNozeEcSG6BpNYJjqU9GHyEZp+GxriR0+6AmYyVkQJwGA3gvIhOXl+896opbi0p1uaSFv/oyDMOuaw2SvHHLgNChpFEXYIwBIri381A9WVUWG+M3nzrWotrYsK4obHBzN6oVvtIU6AgAjBZY4sYcoev3kREbcAwRxEMbJqmdPCiybFz4NMbOzDtweAKi42y1R9Ey4B0VFDSrrDsTsVEHZH5i9mcyDcBoA8P+RV0PR2f8ZIyMgnFXPvrq6O6sXvqGF0HcEALX9SbPgkCeDWZcJ94GAcKIxscLjjarsWodoWvrumjOZAxDPuHWrK+tLgRno4idbEv5kdFx1rs443lbOjZFoi7CLd7pEom3y+ysiznLPby41dQDV1T1CAy0uzr7eaNneFDRVALjXPqy2f7Mroqm7X++ph73+9qEJCCdSEcjGR5bNrDVBbDYn19vXCMDxrIX0BPupqqpTXbhQqU2ffOqDgyMvr0gVFt4J/OFjPc7okWN5TZ4eV339wB886JGuSEcAkDgAnL3UHqKaLEN//4KP7ycB4bIE5M7oNm2WRbPx0SfwUjyEzU6ci4MOjgNBTqTP0ICUgw8izLFjj7T5Vqo3wwNtAvWqkZERjYSdqry8XOXk5Kj29m71wgsVqqCgUPuKtfL7/v5+de3atQAg3PX1wq/M4qagqQAALd7D+yYw6ACdLXyI5rkp0SKgwAi15PE3ltM2d+pZCOQicEOF65B+Dqh9s/H5Gddldr8KEQQBAPLzTwoAnD59WhUVFcnmHh0d1W9aqi2BK7LRa2pqNEK2aKugUJ07d04NDw/Lv8ePnxNlYT/y6Y+o0MkDAKZvJADws14gzUEIDMf7euTZkY0f/e6Sal2YtlWiLNogSAlpCcuVADapfDY7HAayOZzs8bv0QfP+hYCJzygTE7+np0cjSoX8DkuAk58Nf+xYYRAQGHl5+WI6meYF6ENcgZOVMlsvAwNy4OfOtRwBQBIAwGJf/2w7ZPPTTu30C/4FVLJit255IzDORj97tlnUqija47DiRKd4rbl5QHx4wAo5v7y8ZAOwB35Qq8cJsQLGxsbUiRMn9PfnxRIYHx8PxAlu61GkEei6amtrC7gBgTpxjTr4UKATaAo6kVpBbw1goOaaSWMxeCHomO21EOagCCYZVwi5td7deekZiOoSNfqYxH6sD2Cgju12j8Dzl9pV1+pTNbQ3p5p6jH3CfuFARa+TZ0O2jn1l/545sPnPqNzcPDH1cQmMYF+FuAYAgPH9Qz1K9IKp1BuoK/Czw/0Tes9xIUhwt7UNibUAOMC/5iKpxeYUAtHcshr6+0cj+EVHAJDswCQe/s6iAIHXo+GRBqftw4f2cxqQHUd9i3tMQA6Li809ovfUzhc7avrlBTW0PqM2P99WVfpZuLcmwzZ/Tk6eamhoUHV1dfK14RrcCrMAyAwUR7QAkjFzsAawCkBefBjTpwEBMdGxJDDLCGKQogEF7SrfTYb4cQQAhzzTggYRX0GDwW/3ATIUazC5ytJ6cXl5D/xwDjrWL2Y7PQdMHx0A4F7jet5r0QehpS1drd5jjdOTLl5zcPOf1ZsqV9XX1wsA8PWxY6clQHjs2IthMYDikBiAkQ60f3JYAdxQrIJIN7Snx4hwmqImNLYAZeNVOM5K6qcLAGCy4+heRANTP90HtCExwSMJy7AeWV8E2CCQYSnw2s5O48DiMOHwIpNADIxDDbMdqfFoBxYMSnQJTJk6MgmPR8bcBoAHssFJ8WH6c8rn5ubK4OfHjlUGN3xTU5NGugJtpp8NZgEMoHD3QWFS4buz6eEyAALceMwqLIcRfROR+gIwAA6sC/TeeMCnTjXJw2QQUDkCAPsBgEEfQ/LkD/u8nRZkc2JVctiwllg3VVVdYqYTbOPk5sDBZWWD00WJABxxLtza0tLGpK1IXCao1LPvrKmJN1ZEqKTY1ZhUAABKSkrExA8fN27c0L+/IoHA/Px9HsDAwICneQA8EOIOIDBIjAWBK0Hhj+licPoT6AQ0AA9IQZhnPFhQHteEEtGCgvojAEhy0EMQye5bDf1pWQP40riNbFSuj3XAyc3zJg5FDIh1gFXJOmDDszYIBl650iEHBgdN8lH2eACoXl2u7hYuQb7rGTP538MQBuDBcSUjmYAgfnf3iAQdWSAEgFggAAEoj/UA+rMgzBOARYLbwemA1cFJgfsBYODTsVg4TZxcME4MrCQWvCNCG9qFQ46dbszJb+YG2cyY4AAz95uejgSOeV48D8xxng8nOJsasRui6capbQA8zxdLEMIMzyvcPMd9zK6S4Cw2exPpCU+hC/Xi5GUhDnGasPFZgJiDLDAAg7QiQDE6alhQ/Mv3/Jzf8zpOW/4GP9EEEKwU3pP3LitrkcXJyXVSm9HQOflsWJu5ud4GAICPeTIAQgCR67j9uFdtfrajXtTAynVyvZjbWFzcB8CUTQy4khUCeLlvJvji0nFaU7vCfcSSY86mxUacCBAHzLH8kg0S85yzKy6UxQDgRk8ENoS5EdjUVvAgswEAAAa4JwADG6C5eVDMU04uNgInmnUzACxsCL436Z5sDF5vHWwU3iuewfvzWeE/N+dhDnM+1jmZw9yokebFXAbmp9TWr7dVS8+IXC/XDYByH0I3cbO4Xvug5641xTW4asHlGAKr9CkoPdNyBABuDT+rwEY7aa0DwMGaiGcAPLDgwn/OZrS+p2mRWK0Sc8RjndzWZvrKJ5vq5HnvdmICuFwjqenNP/DSgmRMenfmpLS47ErHEQC4MTiFH2SIkEXqLkCf+NRufNb97hG19MG6ang6IeW0fbtzUlmYrmu/oF0ReAtbX+6ozc931NrPN9Stx+6sC66bvgSmlUO/xe6t2SMAcGMQEMKHPAKAxItgUh3LH26E1BDQ7PPMJfdrCJAHp56BBieU8VLfMLwwre4kYRmGuypE9w+L6nPaL/xsPdhq/W7boOrbmzsCAPcKP47EQNwGAFKDkaoIO5/NRP0bCDN087mtXZXw3o4pZYIutsln51lSvbhDxCKCfIZTTVLngHnOaf2wf78r8tovt4W4Q5u0hZ+uB39G81JTKwGrosxyb8Pfb/WTLdX/0ryoK+/9bk+andLOzux7aAWW+okJ6XdJPwY4Fqm3ZsvyRU/q7wgA3AUATO5IADDyauQuQzX9o0KQgSYLa85OijEbH54Cp27l/R5pMhpSGalP5qm3VqTun9eeON8iG+9qoF6AzT786qJR/BQ4xVEFCr+2rS92DOCK8H5LH2yqDg1+SKxvaGsEajAEIQbXezvQrATdQQDDJArR9NbsbHwEAEmMVIQwjwAg+UEn3p2vD24S/N9oqjtmYKywpFFUd3JtrN84ca5FtS8/lRN395s9tfD2M3XvoeECnK/sENVja2ETjWY5mU0AqAzTlMSdiQRwSIbF836Xqvat0obpCQELcy5YDZCG7Lv+LF704abeEQC4Zw1dfdQrp6K5OYiER0v3zb23pi7eNjZFyelmCdY5VWlI38B5DQBz+uSV6kA9z53f7Iqfbg5ai1PtaG7Yc2E6Cju/2Y0IAJzWcb2fRZ69dnQ8pKM1MYKnb68KiPRuz9lAG87iRe9lMZDT5a1q7PUl8R8N26t4AAAB1UlEQVThinN6ZBIAiPmtzWY2z2Fy2lX6OT370y05JTmlG6acrZZrHRuXslwzSMemPBaFWBS+YRmbnx8EgJ2v95J6v3AACOpunGhSw68sqtbF6SMASF4MZNizYiCkpQgulZxqlkXgtNZcOgAgUToxUmPlNqv14Ec3Pp1Ufc/n1dKHG9LlaOm9dTX3w5VgJJ+AXtPspLguWB6k7szgWyQA4Gerv9gSV2BNm+wEA5vmJpN6PysAnNXuKtePpYQL0Lc7L63OjwAgyQFPHC64F+e2/fW+uiyBqe0vd7MaAJwaxaVNav79dYkrsGExrXe+2lNNliahWCiY22xOLLLpH64G/f6IAKA3/bq2IJ5pENj4bEdGhUVlKJH3swLAhRudUjVIkJEeBGQOikqOXICkB3RVr1b6sSgIimEmk3aiC6+Tn0eFXLZmREi5XbSkg9f1JmyPEpCMC7w1mCASan6PJkLL/LRHrz+LAYCiD7uUhWyvVCxvk4VJXphUEDGBIwBwZvRszanm2SnxywECcuztnckpG5ff7NLP7Lm6dHv/xD9/pV0tf7RxBABe49KPjByJgRwBgJFZwMLi5KZs+W7DgMSHEmYCar987t01AZBwt4D4Amk8r137/w+JcpHmAj82BwAAAABJRU5ErkJggg==`
