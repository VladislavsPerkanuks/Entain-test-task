FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache curl

COPY --from=builder /app/main .
COPY migrations ./migrations

CMD ["./main"]
