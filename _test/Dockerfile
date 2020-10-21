FROM golang:1.14-alpine AS builder
COPY . /build
RUN apk add make
WORKDIR /build
RUN make

FROM alpine:3.10
COPY --from=builder /build/_test /dist/_test
WORKDIR /dist
ENTRYPOINT ["/dist/_test"]
