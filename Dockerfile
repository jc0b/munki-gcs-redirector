FROM golang:1.20-alpine AS build_base

WORKDIR /build/munki-gcs-redirector

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY redirector.go .

RUN go build -o ./out/ .

FROM alpine:3.18
RUN apk add ca-certificates

ENV AUTH_USERNAME=bob
ENV AUTH_PASSWORD=bobspassword
ENV GCS_BUCKET_NAME=some_bucket

COPY --from=build_base /build/munki-gcs-redirector/out/munki-gcs-redirector /app/munki-gcs-redirector

EXPOSE 8080

CMD ["/app/munki-gcs-redirector"]