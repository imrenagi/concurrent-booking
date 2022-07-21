FROM golang:1.18 as golang
RUN mkdir -p /
WORKDIR /
COPY . .
RUN make build

FROM alpine:3 as alpine
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates

FROM alpine:3
ENTRYPOINT []
WORKDIR /
COPY --from=alpine /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine /etc/passwd /etc/passwd
COPY --from=golang /bin/booking .