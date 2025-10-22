package main

import (
	"context"
	"fmt"
	"log"
	pb "mockPayment/mock_payment"
	"net"
	"sync"
	"time"

	internal "mockPayment/payment_server/internal"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPaymentServiceServer
	exchangeRateParser *internal.MockParser
	db                 *internal.DB
}

func (s *server) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	providerName, amount, date := req.GetProviderName(), req.GetAmount(), req.GetDate()

	log.Printf("Получен запрос на платёж: %s, %s, %s",
		providerName, amount, date)

	roubleAmount, err := internal.ValidateRequest(s.exchangeRateParser.ExchangeRate, providerName, amount, date)
	if err != nil {
		return &pb.PaymentResponse{Accepted: false}, err
	}

	id, err := s.db.AddPayment(providerName, roubleAmount, date)
	if err != nil {
		return &pb.PaymentResponse{Accepted: false}, err
	}

	return &pb.PaymentResponse{Accepted: true, Id: id, Message: fmt.Sprintf("Платеж принят в обработку. К оплате: %v рублей", roubleAmount)}, nil
}

func main() {
	// Костыльное ожидание запуска БД
	//time.Sleep(20 * time.Second)
	// Инициализация БД, мока
	db, err := internal.NewDB()
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()
	log.Println("Успешно подключились к БД")

	mockERParser := internal.NewMockParser()
	myServer := &server{
		exchangeRateParser: mockERParser,
		db:                 db,
	}

	log.Print("Получение актуального курса валют...")
	err = myServer.exchangeRateParser.UpdateExchangeRate()
	if err != nil {
		log.Fatalf("Ошибка при получении актуального курса валют: %v", err)
	} else {
		log.Print("Актуальный курс валют получен")
	}

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, myServer)
	log.Printf("сервер запущен на %v", lis.Addr())

	var wg sync.WaitGroup

	// Горутина для удаления просроченных записей
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				affected, err := myServer.db.DeleteExpiredPayments()
				if err != nil {
					log.Printf("Ошибка при удалении просроченных записей: %v", err)
				} else if affected > 0 {
					log.Printf("Удалено %d просроченных записей", affected)
				}
			}
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
	}

	wg.Wait()
}
