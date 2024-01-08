FROM docker.io/golang:1.20-alpine3.19 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY ui ./ui
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s' -o /web ./cmd/web

FROM docker.io/alpine:3.19 AS build-release-stage
WORKDIR /
COPY --from=build-stage /web /web
EXPOSE 4000
ENTRYPOINT ["/web"]
