.PHONY=dist clean
BUILDARGS=-ldflags='-w -s' -trimpath

all:
	CGO_ENABLED=0 GOOS=linux go build -o _test

image:
	docker build . -t ghcr.io/jdevelop/repull:latest
	docker push ghcr.io/jdevelop/repull:latest
