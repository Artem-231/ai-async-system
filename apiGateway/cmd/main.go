package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Artem-231/ai-async-system/gateway/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	internal.InitDB()
	defer internal.DB.Close()

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	internal.Channel, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer internal.Channel.Close()

	internal.Queue, err = internal.Channel.QueueDeclare(
		"task_queue", true, false, false, false, nil,
	)
	failOnError(err, "Failed to declare a queue")

	mux := http.NewServeMux()
	mux.HandleFunc("/task", internal.HandleTask)
	mux.HandleFunc("/status", internal.HandleStatus)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Printf("Gateway is running on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Получен сигнал на остановку. Завершаем работу Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Ошибка при остановке сервера:", err)
	}

	log.Println("Gateway успешно остановлен без потери данных")
}

// failOnError облегчает просмотр кода, благодаря вынесу вывода ошибка в отдельную функцию
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
