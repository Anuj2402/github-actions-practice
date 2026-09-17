# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod ./
RUN go mod download

COPY main.go ./
COPY index.html ./
COPY history.html ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/chat-server main.go

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup \
    && adduser -S -G appgroup -H -h /app appuser

WORKDIR /app

COPY --from=builder /out/chat-server ./chat-server
COPY --from=builder /src/index.html ./index.html
COPY --from=builder /src/history.html ./history.html

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --spider http://127.0.0.1:8080/ || exit 1

ENTRYPOINT ["./chat-server"]
