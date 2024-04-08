FROM golang:1.22.2-alpine AS build

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY *.go .
COPY migrations/*.g[o] ./migrations/

RUN CGO_ENABLED=0 go build -o pb

FROM alpine:3

RUN apk add --no-cache \
    unzip \
    ca-certificates

WORKDIR /pb

COPY --from=build /build/pb .

COPY ./migration[s] ./pb_migrations
COPY ./hook[s] ./pb_hooks

CMD ["sh", "-c", "/pb/pb serve --http=0.0.0.0:${PORT:-8080}"]