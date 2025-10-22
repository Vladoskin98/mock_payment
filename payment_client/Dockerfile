FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY ../go.mod ../go.sum ./
COPY mock_payment/ ./mock_payment/
COPY payment_client/ ./payment_client/

RUN go mod download
RUN go build -o /app/payment-client ./payment_client/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/payment-client .
CMD ["./payment-client"]