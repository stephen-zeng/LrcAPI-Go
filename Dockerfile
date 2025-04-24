FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o lrcAPI .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/lrcAPI .

EXPOSE 1111

CMD ["./lrcAPI"]
