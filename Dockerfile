FROM golang:latest as builder

WORKDIR /build
ADD . /build

RUN CGO_ENABLED=0 go build -o simple-server ./cmd/

FROM alpine:latest
LABEL org.opencontainers.image.source=https://github.com/sturtevant/simple-server/

WORKDIR /svc
COPY --from=builder /build/simple-server /svc/
ENTRYPOINT ["/svc/simple-server"]
