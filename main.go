package main

import (
	"encoding/json"
	"fmt"
	//"io"
	"net/http"
	"os"
	"bytes"
	"image/png"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

// This handler is called every time Telegram sends us a webhook event
func Handler(res http.ResponseWriter, req *http.Request) {
	// Initialize bot with token
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		fmt.Println("error creating bot:", err)
		return
	}

	// First, decode the JSON response body
	body := &tgbotapi.Update{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		fmt.Println("could not decode request body", err)
		return
	}

	// Check if the update has a message
	if body.Message == nil {
		return
	}

	// Check if the message contains text
	if body.Message.Text == "" {
		return
	}

	code := body.Message.Text

	if err := sendMessage(bot, body.Message.Chat.ID, code); err != nil {
		fmt.Println("error in sending reply:", err)
		return
	}

	// log a confirmation message if the message is sent successfully
	fmt.Printf("Reply sent: %q \n", body.Message.Text)
	fmt.Printf("Reply sent to: %q", body.Message.From.UserName)
}

func createImg(codeText string) []byte {
	code, err := code128.Encode(codeText)
	if err != nil {
		fmt.Println("error creating code128: ", err)
	}

	// Create image
	codeImg, err := barcode.Scale(code, 250, 100)
	if err != nil {
		fmt.Println("error creating img: ", err)
	}

	// Create buffer in memory
	buf := new(bytes.Buffer)

	// Encode image into PNG and write it to the buffer
	err = png.Encode(buf, codeImg)
	if err != nil {
		fmt.Println("Ошибка при кодировании изображения: ", err)
	}

	// Convert the buffer to []bytes
	imageBytes := buf.Bytes()

	return imageBytes
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, code string) error {
	// Create image from text
	imageBytes := createImg(code)
	bytes := tgbotapi.FileBytes{Name: "image.png", Bytes: imageBytes}

	// Preparing message with image
	msg := tgbotapi.NewPhoto(chatID, bytes)

	// Add keyboard with buttons
	msg.ReplyMarkup = addKeyboard()

	// Send message with keyboard
	if _, err := bot.Send(msg); err != nil {
		fmt.Printf("Error sending photo: %s", err)
	}

	return nil
}

// Create keyboard with buttons
func addKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("V-Sales_814"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Discount_814"),
			tgbotapi.NewKeyboardButton("V-Discount_814"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("010_814-Exit_Dost"),
			tgbotapi.NewKeyboardButton("010_814-01-02-3"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("010_814-Exit_sklad"),
			tgbotapi.NewKeyboardButton("010_814-Exit_zal"),
		),
	)
	return keyboard
}

func main() {
	// Create a new HTTP server
	// Start the server on port 3000
	http.ListenAndServe(":3000", http.HandlerFunc(Handler))
}
