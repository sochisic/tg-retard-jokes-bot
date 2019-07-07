package main

import (
	"fmt"
	"net/http"
	"os"
	"tg-retards-joke-bot/pictures"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type User struct {
	ID           int
	UserName     string
	FirstName    string
	LastName     string
	SeenJokes    []string
	JokesExpires time.Time
	AlreadyBeen  bool
	ChatArchive  []string
}

type Users map[int]*User

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

var sublogger = log.With().Str("component", "pictures").Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})

var users = Users{}
var pics = pictures.Pictures{Logger: &sublogger}

func main() {
	BotToken, exists := os.LookupEnv("TG_BOT_TOKEN")
	if !exists {
		log.Fatal().Msg("tg token is required")
	}

	WebhookURL, exists := os.LookupEnv("WEBHOOK_URL")
	if !exists {
		log.Fatal().Msg("WebhookURL is required")
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	debug, exists := os.LookupEnv("DEBUG")
	if !exists {
		log.Print("DEBUG Env is missing set log level to INFO")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	if debug == "true" {
		log.Print("DEBUG Env variable 'true' set log level to DEBUG")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		log.Print("DEBUG Env variable not 'true' set log level to INFO")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Print("Hello World")

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}

	// bot.Debug = true
	log.Info().Str("Authorized on account:", bot.Self.UserName).Send()

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		panic(err)
	}

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(":8080", nil)
	log.Info().Msg("start listen :8080")

	for update := range updates {
		log.Debug().Msgf("[%s] %s \n", update.Message.From.UserName, update.Message.Text)
		welcomeMessage := "Псссст, я смотрю ты первый раз тут, хочешь немного приколов для даунов?"

		if _, ok := users[update.Message.From.ID]; ok {
			users[update.Message.From.ID].ChatArchive = append(users[update.Message.From.ID].ChatArchive, update.Message.Text)
			welcomeMessage = fmt.Sprintf("О привет %s ты вернулся, хочешь ещё приколов для даунов?", update.Message.From.UserName)
		} else {
			users[update.Message.From.ID] = &User{
				ID:           update.Message.From.ID,
				AlreadyBeen:  true,
				JokesExpires: time.Now().Add(240 * time.Hour),
				UserName:     update.Message.From.UserName,
				FirstName:    update.Message.From.FirstName,
				LastName:     update.Message.From.LastName,
				ChatArchive:  []string{update.Message.Text},
			}
		}

		switch update.Message.Text {

		case "да", "Да", "yes", "Yes", "y", "д":
			if _, ok := users[update.Message.From.ID]; ok {

				pic, err := getNotSeenPicture(users[update.Message.From.ID].SeenJokes, update.Message.From.ID)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Случилась неудача, попробуй ещё раз",
					))
				}

				_, error := bot.Send(tgbotapi.NewPhotoShare(update.Message.Chat.ID, pic))
				if error != nil {
					bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Случилась неудача, попробуй ещё раз",
					))
				}
				users[update.Message.From.ID].SeenJokes = append(users[update.Message.From.ID].SeenJokes, pic)
			} else {
				pic, error := pics.GetPicture(update.Message.From.ID)
				if error != nil {
					bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Случилась неудача, попробуй ещё раз",
					))
				}

				_, err := bot.Send(tgbotapi.NewPhotoShare(update.Message.Chat.ID, pic))
				if err != nil {
					bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Случилась неудача, попробуй ещё раз",
					))
				}
			}

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"хочешь ещё?",
			))

		case "нет", "н", "no", "No", "Нет":
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Возвращайся когда захочешь",
			))
		default:
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				welcomeMessage,
			))

			for k, v := range users {
				log.Debug().Msgf("key[%v] value[%v]\n", k, v.UserName)
				log.Debug().Msgf("key[%v] value[%v]\n", k, v.FirstName)
				log.Debug().Msgf("key[%v] value[%v]\n", k, v.LastName)
				log.Debug().Msgf("key[%v] value[%v]\n", k, v.SeenJokes)
			}

		}
	}
}

func getNotSeenPicture(seen []string, id int) (string, error) {
	pic, err := pics.GetPicture(id)

	if err != nil {
		return "", err
	}

	for contains(seen, pic) {
		pic, err = pics.GetPicture(id)
	}

	return pic, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
