package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/opesun/goquery"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const url = "http://joyreactor.cc/tag/%23%D0%9F%D1%80%D0%B8%D0%BA%D0%BE%D0%BB%D1%8B+%D0%B4%D0%BB%D1%8F+%D0%B4%D0%B0%D1%83%D0%BD%D0%BE%D0%B2"

type Pictures struct {
	Items     []string
	expiresAt time.Time
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	BotToken, exists := os.LookupEnv("TG_BOT_TOKEN")
	if !exists {
		log.Fatal("tg token is required")
	}

	WebhookURL, exists := os.LookupEnv("WEBHOOK_URL")
	if !exists {
		log.Fatal("WebhookURL is required")
	}

	sessions := make(map[int64]int)
	pictures := Pictures{}

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}

	// bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		panic(err)
	}

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(":8080", nil)
	fmt.Println("start listen :8080")

	for update := range updates {
		fmt.Printf("[%s] %s \n", update.Message.From.UserName, update.Message.Text)
		switch update.Message.Text {
		case "да", "Да", "yes", "Yes", "y", "д":
			x, err := goquery.ParseUrl(url)
			if err != nil {
				panic(err)
			}

			if len(pictures.Items) == 0 {
				fmt.Println("pictures not found, updating pictures array")
				pictures.Items = x.Find("#post_list .postContainer .article div.post_top div.post_content div.image img").Attrs("src")
				pictures.expiresAt = expiresIn15min()
			}

			if time.Now().After(pictures.expiresAt) {
				fmt.Println("pictures is expired, updating pictures array")
				pictures.Items = x.Find("#post_list .postContainer .article div.post_top div.post_content div.image img").Attrs("src")
				pictures.expiresAt = expiresIn15min()
			}

			if len(pictures.Items) != 0 {
				if val, ok := sessions[update.Message.Chat.ID]; ok {
					_, err := bot.Send(tgbotapi.NewPhotoShare(update.Message.Chat.ID, pictures.Items[val]))
					if err != nil {
						bot.Send(tgbotapi.NewMessage(
							update.Message.Chat.ID,
							"Случилась неудача, попробуй ещё раз",
						))
					}

					sessions[update.Message.Chat.ID] = sessions[update.Message.Chat.ID] + 1
				} else {
					_, err := bot.Send(tgbotapi.NewPhotoShare(update.Message.Chat.ID, pictures.Items[val]))
					if err != nil {
						bot.Send(tgbotapi.NewMessage(
							update.Message.Chat.ID,
							"Случилась неудача, попробуй ещё раз",
						))
					}

					sessions[update.Message.Chat.ID] = 1
				}

				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"хочешь ещё? ",
				))

			} else {
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Нету приколов для даунов :/",
				))
			}
		case "нет":
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Возвращайся когда захочешь",
			))
		default:
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Хочешь немного приколов для даунов?`,
			))
		}
	}
}

func expiresIn15min() time.Time {
	return time.Now().Add(15 * time.Minute)
}
