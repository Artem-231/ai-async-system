package main

import (
	"log"
	"net/http"

	"github.com/Artem-231/ai-async-system/gateway/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
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

	http.HandleFunc("/task", internal.HandleTask)
	http.HandleFunc("/status", internal.HandleStatus)

	log.Printf("Gateway is running on http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// failOnError облегчает просмотр кода, благодаря вынесу вывода ошибка в отдельную функцию
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
