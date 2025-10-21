package main

import (
	"context"
	"fmt"
	"log"
	pb "mockPayment/mock_payment"
	"net"

	internal "mockPayment/payment_server/internal"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPaymentServiceServer
	ActualExchangeRate map[string]float64
}

func (s *server) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	providerName, amount, date := req.GetProviderName(), req.GetAmount(), req.GetDate()

	log.Printf("Получен запрос на платёж: %s, %s, %s",
		providerName, amount, date)

	roubleAmount, err := internal.ValidateRequest(s.ActualExchangeRate, providerName, amount, date)
	if err != nil {
		return &pb.PaymentResponse{Accepted: false}, err
	}

	return &pb.PaymentResponse{Accepted: true, Id: 1, Message: fmt.Sprintf("Платеж принят в обработку. К оплате: %v рублей", roubleAmount)}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	s := grpc.NewServer()
	mockERParser := internal.MockParser{}
	ActualExchangeRate, err := mockERParser.GetActualExchangeRate()
	if err != nil {
		log.Fatalf("Ошибка при получении актуального курса валют: %v", err)
	}

	pb.RegisterPaymentServiceServer(s, &server{ActualExchangeRate: ActualExchangeRate})
	log.Printf("сервер запущен на %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
	}
}
