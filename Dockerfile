FROM golang:1.17.0-alpine3.14 AS builder
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

