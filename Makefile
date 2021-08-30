.PHONY=all test
BUILDARGS=-ldflags='-w -s' -trimpath

all: test
	CGO_ENABLED=0 GOOS=linux go build ${BUILDARGS}

test:
	go test -v ./...

install:
	go install ${BUILDARGS}

docker: test
	docker build -t ghcr.io/jdevelop/repull:latest .
