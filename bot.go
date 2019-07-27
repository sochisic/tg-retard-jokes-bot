package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sochisic/tg-retard-jokes-bot/pictures"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

//User struct it is representation of bot visitor
type User struct {
	UserID       int
	ChatID       int64
	UserName     string
	FirstName    string
	LastName     string
	SeenJokes    []string
	JokesExpires time.Time
	ChatArchive  []string
}

//Users it is map of visitors where key it is ID of user
type Users map[int]*User

//ChatMessage struct represent message to bot undepend of text or callback message it
type ChatMessage struct {
	UserID         int
	ChatID         int64
	Name           string
	Text           string
	WelcomeMessage string
	IsCommand      bool
}

//YesOrNoOrTiredKeyboard Inline Keyboard with three button 'Да', 'No' and 'Я устал' as well
var YesOrNoOrTiredKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("да", "да"),
		tgbotapi.NewInlineKeyboardButtonData("нет", "нет"),
		tgbotapi.NewInlineKeyboardButtonData("я устал", "/tired"),
	),
)

//YesKeyboard it is Inline Keyboard for answer to chat with button 'Yes'
var YesKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("да", "да"),
	),
)

//ReturnKeyboard it is Inline Keyboard for answer in chat with only one button 'Я вернулся'
var ReturnKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Я вернулся", "/start"),
	),
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("No .env file found")
	}
}

var sublogger = log.With().Str("component", "pictures").Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})

var users = Users{}
var pics = pictures.Pictures{Logger: &sublogger}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	BotToken, exists := os.LookupEnv("TG_BOT_TOKEN")
	if !exists {
		log.Fatal().Msg("tg token is required")
	}

	WebhookURL, exists := os.LookupEnv("WEBHOOK_URL")
	if !exists {
		log.Fatal().Msg("WebhookURL is required")
	}

	WebhookPort, exists := os.LookupEnv("WEBHOOK_PORT")
	if !exists {
		log.Warn().Str("WEBHOOK_PORT env variable not defined! Using default port", "8080").Send()
		WebhookPort = "8080"
	}

	debug, _ := os.LookupEnv("DEBUG")
	if debug == "true" {
		log.Info().Str("DEBUG Env variable 'true' set log level to", "DEBUG").Send()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		log.Info().Str("DEBUG Env variable missing or 'false' set log level to", "INFO").Send()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Panic().Err(err).Send()
		panic(err)
	}

	botDebug, _ := os.LookupEnv("BOT_DEBUG")
	if botDebug == "true" {
		log.Print("DEBUG Env variable 'true' set log level to DEBUG")
		bot.Debug = true
	}

	log.Info().Str("Authorized on account", bot.Self.UserName).Send()

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		panic(err)
	}

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(":"+WebhookPort, nil)
	log.Info().Str("Start listen port", WebhookPort).Send()

	for update := range updates {
		message := ChatMessage{
			WelcomeMessage: "Псссст, я смотрю ты первый раз тут, хочешь немного приколов для даунов?",
		}

		var userName string
		var lastName string
		var firstName string

		messageExist := update.CallbackQuery != nil || update.Message != nil

		if update.Message != nil {
			message.UserID = update.Message.From.ID
			message.Text = update.Message.Text
			message.ChatID = update.Message.Chat.ID
			message.IsCommand = update.Message.IsCommand()
			if len(update.Message.From.UserName) != 0 {
				message.Name = update.Message.From.UserName
			} else if len(update.Message.From.FirstName) != 0 {
				message.Name = update.Message.From.FirstName
			} else if len(update.Message.From.LastName) != 0 {
				message.Name = update.Message.From.LastName
			} else {
				message.Name = string(update.Message.From.ID)
			}

			userName = update.Message.From.UserName
			lastName = update.Message.From.LastName
			firstName = update.Message.From.FirstName
		}

		if update.CallbackQuery != nil {
			message.UserID = update.CallbackQuery.From.ID
			message.Text = update.CallbackQuery.Data
			message.ChatID = update.CallbackQuery.Message.Chat.ID
			if update.CallbackQuery.Data == "/start" || update.CallbackQuery.Data == "/tired" {
				message.IsCommand = true
			}

			if len(update.CallbackQuery.From.UserName) != 0 {
				message.Name = update.CallbackQuery.From.UserName
			} else if len(update.CallbackQuery.From.FirstName) != 0 {
				message.Name = update.CallbackQuery.From.FirstName
			} else if len(update.CallbackQuery.From.LastName) != 0 {
				message.Name = update.CallbackQuery.From.LastName
			} else {
				message.Name = string(update.CallbackQuery.From.ID)
			}

			userName = update.CallbackQuery.From.UserName
			lastName = update.CallbackQuery.From.LastName
			firstName = update.CallbackQuery.From.FirstName
		}

		if messageExist {
			log.Debug().Str("Message", message.Text).Int("From ID", message.UserID).Str("UserName", message.Name).Send()

			if _, ok := users[message.UserID]; ok {
				users[message.UserID].ChatArchive = append(users[message.UserID].ChatArchive, message.Text)

				message.WelcomeMessage = fmt.Sprintf("О привет %s ты вернулся, хочешь ещё приколов для даунов?", message.Name)
			} else {
				users[message.UserID] = &User{
					UserID:       message.UserID,
					ChatID:       message.ChatID,
					JokesExpires: time.Now().Add(240 * time.Hour),
					UserName:     userName,
					FirstName:    firstName,
					LastName:     lastName,
					ChatArchive:  []string{message.Text},
				}
			}

			if message.IsCommand {
				msg := tgbotapi.NewMessage(message.ChatID, message.Text)

				switch message.Text {
				case "/start":
					msg.ReplyMarkup = YesKeyboard
					msg.Text = message.WelcomeMessage
				case "/tired":
					msg.Text = "Тут пока ничего нет :/"
				}

				bot.Send(msg)

			} else {
				chatModule(message, bot)
			}
		}
	}
}

func getNotSeenPicture(seen []string, id int) (string, error) {
	pic, err := pics.GetPicture(id)
	if err != nil {
		log.Error().Err(err).Send()
		return "", err
	}

	for contains(seen, pic) {
		pic, err = pics.GetPicture(id)
		if err != nil {
			log.Error().Err(err).Send()
			return "", err
		}
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

func chatModule(m ChatMessage, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(m.ChatID, m.Text)

	switch m.Text {
	case "да", "Да", "yes", "Yes", "y", "д":
		pic, err := getNotSeenPicture(users[m.UserID].SeenJokes, m.UserID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(
				m.ChatID,
				"Случилась неудача, попробуй ещё раз",
			))
		}

		_, error := bot.Send(tgbotapi.NewPhotoShare(m.ChatID, pic))
		if error != nil {
			bot.Send(tgbotapi.NewMessage(
				m.ChatID,
				"Случилась неудача, попробуй ещё раз",
			))
		}

		msg.ReplyMarkup = YesOrNoOrTiredKeyboard
		msg.Text = "хочешь ещё?"

		users[m.UserID].SeenJokes = append(users[m.UserID].SeenJokes, pic)
	case "нет", "н", "no", "No", "Нет":
		msg.Text = "Возвращайся когда захочешь..."
		msg.ReplyMarkup = ReturnKeyboard
	default:
		msg.Text = "Неизвестная команда, но я могу прислать тебе немного приколов для даунов, хочешь?"
		msg.ReplyMarkup = YesKeyboard
	}

	bot.Send(msg)
}
