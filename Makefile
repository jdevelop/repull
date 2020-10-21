.PHONY=dist clean
BUILDARGS=-ldflags='-w -s' -trimpath

all:
	CGO_ENABLED=0 GOOS=linux go build ${BUILDARGS}
