all: fmt bench coverage doc

prepare-dev:
	go get github.com/axw/gocov/gocov	
	go get gopkg.in/matm/v1/gocov-html

presubmit: fmt
	go test github.com/slspeek/apod-bg...

doc:
	mkdir -p report/doc
	godoc -html  github.com/slspeek/apod-bg/apod > report/doc/index.html

fmt:
	go fmt github.com/slspeek/apod-bg... 

bench:
	go test -benchmem -bench=. github.com/slspeek/apod-bg... 

coverage:
	mkdir -p report
	gocov test github.com/slspeek/apod-bg/apod | gocov-html > report/coverage-apod.html

