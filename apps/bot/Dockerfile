FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/bot ./cmd/bot

FROM alpine:3.20

RUN adduser -D -g '' appuser
USER appuser
WORKDIR /app
COPY --from=builder /out/bot /app/bot

CMD ["/app/bot"]
