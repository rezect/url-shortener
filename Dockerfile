FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o backend ./cmd/server/main.go

FROM alpine:latest

WORKDIR /home/appuser

COPY --from=builder /app/backend .
COPY /migrations ./migrations

RUN chmod +x /home/appuser/backend

EXPOSE 8000

CMD ["./backend"]