#build stage
FROM golang:1.22.5-alpine AS builder

WORKDIR /app
ENV CGO_ENABLED 0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -tags=production -ldflags="-s -w" -o /app/build/goserver main.go

#final stage
FROM alpine:latest
RUN apk update && apk add supervisor
RUN adduser -D appuser
USER appuser
WORKDIR /app/
COPY --from=builder /app/build .

ARG APP_PORT=8415
EXPOSE $APP_PORT

USER root
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
CMD ["supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
