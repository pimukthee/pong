FROM golang:1.21.1-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o pong ./cmd/main.go

RUN chmod +x /app/pong


FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/pong /app

COPY --from=builder /app/web /app/web

WORKDIR /app

EXPOSE 8080

CMD ["/app/pong"]
