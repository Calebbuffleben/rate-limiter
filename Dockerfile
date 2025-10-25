# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/server

FROM alpine:3.20
RUN adduser -D -H appuser
WORKDIR /home/appuser
COPY --from=builder /server ./server
ENV PORT=8080
EXPOSE 8080
USER appuser
ENTRYPOINT ["./server"]


