all: fmt test vet bench coverage

clean:
	go clean
	rm -rf report

prepare-dev:
	go get github.com/axw/gocov/gocov	
	go get gopkg.in/matm/v1/gocov-html

presubmit: fmt test vet

test:
	go test github.com/slspeek/apod-bg/apod

vet:
	go vet github.com/slspeek/apod-bg/apod && go vet main.go

fmt:
	go fmt github.com/slspeek/apod-bg... 

bench:
	go test -benchmem -bench=. github.com/slspeek/apod-bg 

coverage:
	mkdir -p report
	gocov test github.com/slspeek/apod-bg/apod | gocov-html > report/coverage-apod.html
