# wyrd_web
# Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
# See LICENSE for distribution and usage details.

FROM golang:1.24-alpine AS builder
RUN apk add build-base

WORKDIR /build

COPY go.mod go.sum /build/
RUN go mod download

COPY . /build

VOLUME [ "/root/.cache" ]
RUN --mount=type=cache,target=/root/.cache/go-build go build -o main -buildvcs=false .

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser

COPY --from=builder /build/main /app/
COPY openapi.yaml /app 
COPY migrations/ /app/migrations/

WORKDIR /app

EXPOSE 8080
CMD [ "/app/main" ]