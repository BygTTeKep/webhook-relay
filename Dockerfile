

FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


ARG SERVICE=api


RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /bin/app \
    ./cmd/${SERVICE}

FROM alpine:3.20

# RUN apk add --no-cache ca-certificates && \
# adduser -D -u 1000 appuser


WORKDIR /app
COPY --from=builder /bin/app .

USER appuser

ENTRYPOINT ["./app"]