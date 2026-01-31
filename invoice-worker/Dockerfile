FROM golang:1.21-apline AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . . 
RUN CGO_ENABLED=0 GOOS=linux go build -o invoice-worker ./cmd/worker/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /roor/
COPY --from=builder /app/invoice-worker .
EXPOSE 8080
CMD ["./invoice-worker"]