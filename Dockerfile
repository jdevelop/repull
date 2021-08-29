FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
# RUN go build .
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .

FROM scratch
WORKDIR /
COPY --from=builder /build/repull .
ENTRYPOINT ["/repull"]

