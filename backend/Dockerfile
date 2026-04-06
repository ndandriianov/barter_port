FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/app ./cmd/${SERVICE}

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /out/app /app/app
COPY --from=builder /src/migrations /app/migrations
#COPY config /app/config

EXPOSE 8080
ENTRYPOINT ["/app/app"]