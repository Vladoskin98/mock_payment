package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	pb "mockPayment/mock_payment"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("ошибка при установлении соединения: %v", err)
	}
	defer conn.Close()
	client := pb.NewPaymentServiceClient(conn)

	// Создаю списки с различными вводными данными
	providerList := []string{"", ".,.,.", " ", "Netflix", "Steam", "JetBrains", "OpenAI"}
	countList := []string{"", " ", "-10", "100", "20.2525", "100.500", "99.99"}
	currencyList := []string{"AKSA", "USD", "AUD", "BYR", "EUR"}
	dateList := []string{"01.02.2006", "20.10.2025", time.Now().Format("02.01.2006")}

	var wg sync.WaitGroup
	for _, provider := range providerList {
		for _, count := range countList {
			for _, currency := range currencyList {
				for _, date := range dateList {
					amount := strings.Join([]string{count, currency}, " ")
					wg.Add(1)
					go func() {
						defer wg.Done()
						ctx, cancel := context.WithTimeout(context.Background(), time.Second)
						defer cancel()
						resp, err := client.ProcessPayment(ctx, &pb.PaymentRequest{ProviderName: provider, Amount: amount, Date: date})
						var sb strings.Builder
						sb.WriteString(fmt.Sprintf("ProviderName: %s, Amount: %s, Date: %s. ", provider, amount, date))
						sb.WriteString(fmt.Sprintf("Accepted: %v", resp.GetAccepted()))
						if err != nil {
							sb.WriteString(fmt.Sprintf(". Ошибка: %s", strings.Replace(err.Error(), "rpc error: code = Unknown desc = ", "", 1)))
						} else {
							sb.WriteString(fmt.Sprintf("; Id: %d; Message: %s.", resp.GetId(), resp.GetMessage()))
						}

						log.Print(sb.String())
					}()

				}
			}
		}
	}
	wg.Wait()
}
