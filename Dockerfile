FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /orbital ./cmd/api

FROM alpine:3.19

RUN addgroup -S orbital && adduser -S orbital -G orbital

WORKDIR /app/run

COPY --from=builder /orbital /usr/local/bin/orbital

USER orbital

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/orbital"]
