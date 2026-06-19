# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o htload ./cmd/htload

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S ogou && adduser -S ogou -G ogou

COPY --from=builder /build/htload /usr/local/bin/htload

USER ogou

ENTRYPOINT ["htload"]
CMD ["--help"]
