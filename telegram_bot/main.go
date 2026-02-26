package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TG_TOKEN")

	if token == "" {
		log.Fatal("TG_TOKEN environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if !update.Message.IsCommand() {
			msg.Text = "Я реагирую только на команды. Нажми /help"
			bot.Send(msg)

			continue
		}

		switch update.Message.Command() {
		case "help":

			msg.Text = "Используйте /generate + {то, что вы хотите нарисовать}"

		case "start":
			msg.Text = "Привет! Я асинхронный ИИ-бот. Напиши /help для списка команд."

		case "generate":
			arg := update.Message.CommandArguments()
			if arg == "" {
				msg.Text = "Пожалуйста, напиши что нарисовать. Пример: /generate киберпанк город"
				break
			}
			req := GatewayRequest{
				Action:  "Generate",
				Payload: update.Message.CommandArguments(),
			}

			reqBytes, err := json.Marshal(req)

			if err != nil {
				log.Panic(err)
			}

			resp, err := http.Post("http://gateway:8080/task", "application/json", bytes.NewBuffer(reqBytes))

			if err != nil {
				msg.Text = "Ошибка связи с сервером"
				break
			}

			if resp.StatusCode != 200 {
				msg.Text = "Ошибка связи с сервером"
				resp.Body.Close()
				break
			}

			var gwResp GatewayResponse
			err = json.NewDecoder(resp.Body).Decode(&gwResp)
			resp.Body.Close()

			if err != nil {
				log.Printf("Ошибка парсинга JSON: %v", err)
				msg.Text = "Ошибка чтения ответа от сервера"
				break
			}

			msg.Text = "Задача отправлена в очередь! Ожидайте..."

			go func(taskID int, chatID int64) {
				for {
					statusResp, err := http.Get(fmt.Sprintf("http://gateway:8080/status?id=%d", taskID))
					if err != nil {
						bot.Send(tgbotapi.NewMessage(chatID, "Потеряна связь с сервером..."))
						return
					}

					var res StatusResponse
					err = json.NewDecoder(statusResp.Body).Decode(&res)
					statusResp.Body.Close()

					if err != nil {
						bot.Send(tgbotapi.NewMessage(chatID, "Ошибка обработки статуса..."))
						return
					}

					switch res.Status {
					case "pending":
						time.Sleep(3 * time.Second)

					case "error":
						bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при генерации картинки!"))
						return

					case "done":
						imagePath := fmt.Sprintf("../images/%d.png", taskID)
						photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))

						bot.Send(photo)
						return
					}
				}
			}(gwResp.ID, update.Message.Chat.ID)

		default:
			msg.Text = "Я реагирую только на команды. Нажми /help"
		}

		bot.Send(msg)

	}
}

type GatewayRequest struct {
	Action  string `json:"action"`
	Payload string `json:"payload"`
}

type GatewayResponse struct {
	ID int `json:"id"`
}

type StatusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
