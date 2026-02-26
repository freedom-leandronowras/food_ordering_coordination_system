# syntax=docker/dockerfile:1.7

FROM golang:1.24.0-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags='-s -w' -o /out/coordination-api .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/coordination-api /app/coordination-api

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/coordination-api"]
