FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY ../go.mod ../go.sum ./
COPY mock_payment/ ./mock_payment/
COPY payment_server/ ./payment_server/

RUN go mod download
RUN go build -o /app/payment-server ./payment_server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/payment-server .
COPY --from=builder /app/payment_server/internal/ ./internal/
EXPOSE 50051
CMD ["./payment-server"]
