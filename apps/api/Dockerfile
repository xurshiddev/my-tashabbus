FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/api ./cmd/api

FROM alpine:3.20

RUN adduser -D -g '' appuser
USER appuser
WORKDIR /app
COPY --from=builder /out/api /app/api

EXPOSE 8080
CMD ["/app/api"]
