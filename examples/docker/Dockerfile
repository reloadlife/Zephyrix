FROM golang:1.23-alpine AS build-env

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o app -ldflags="-w -s" simple-test.go

FROM alpine:3.18

RUN apk add --no-cache tzdata ca-certificates curl
RUN adduser -D appuser
COPY --from=build-env /src/app /usr/local/bin/app
RUN chown appuser:appuser /usr/local/bin/app

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 CMD /usr/local/bin/app health || exit 1

USER appuser
ENTRYPOINT ["/usr/local/bin/app"]
