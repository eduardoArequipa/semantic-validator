FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine3.24 AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod ./
RUN go mod download

COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/semantic-validator ./cmd/server
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/keys ./cmd/keys

FROM alpine:3.24

RUN apk add --no-cache ca-certificates wget \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app
RUN mkdir /data && chown app:app /data
COPY --from=build --chown=app:app /out/keys /app/keys
COPY --from=build --chown=app:app /out/semantic-validator /app/semantic-validator

ENV PORT=8080
EXPOSE 8080
USER app

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null "http://127.0.0.1:${PORT:-8080}/health" || exit 1

ENTRYPOINT ["/app/semantic-validator"]
