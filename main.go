package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/makiuchi-d/gozxing"
	//"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

// const (
// 	lines  int = 19
// 	cells  int = 7
// 	floors int = 2
// )

var (
	line  int
	cell  int
	floor int
)

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

	if body.Message.Text == "Выбрать ячейку" {
		if err := sendCell(bot, body.Message.Chat.ID); err != nil {
			fmt.Println("error in sending cell:", err)
		}
		return
	}

	if strings.HasPrefix(body.Message.Text, "Ряд") || strings.HasPrefix(body.Message.Text, "Ячейка") || strings.HasPrefix(body.Message.Text, "Ярус") {
		str := strings.Split(body.Message.Text, ": ")
		switch str[0] {
		case "Ряд":
			line, _ = strconv.Atoi(str[1])
		case "Ячейка":
			cell, _ = strconv.Atoi(str[1])
		case "Ярус":
			floor, _ = strconv.Atoi(str[1])

		}
		fmt.Printf("Сообщене: %q", str[1])
		if err := sendCell(bot, body.Message.Chat.ID); err != nil {
			fmt.Println("error in sending cell:", err)
		}
		return
	}

	var isASCII bool = true

	for i := 0; i < len(body.Message.Text); i++ {
		if body.Message.Text[i] > unicode.MaxASCII {
			isASCII = false
		}
	}
	if !isASCII {
		sendNotAsciiMessage(bot, body.Message.Chat.ID)
		return
	}

	if err := sendMessage(bot, body.Message.Chat.ID, code); err != nil {
		fmt.Println("error in sending reply:", err)
	}

	// log a confirmation message if the message is sent successfully
	fmt.Printf("Reply sent: %q \n", body.Message.Text)
	fmt.Printf("Reply sent to: %q", body.Message.From.UserName)
}

func sendNotAsciiMessage(bot *tgbotapi.BotAPI, chatID int64) error {
	// Create image from text

	msg := tgbotapi.NewMessage(chatID, "Генератор не поддерживает кириллицу")

	// Send message with keyboard
	if _, err := bot.Send(msg); err != nil {
		fmt.Printf("Error sending photo: %s", err)
	}

	return nil
}

func createImg(codeText string) []byte {
	enc := qrcode.NewQRCodeWriter()
	codeImg, err := enc.Encode(codeText, gozxing.BarcodeFormat_QR_CODE, 256, 256, nil)
	if err != nil {
		fmt.Println("error creating QR: ", err)
	}

	// Create buffer in memory
	buf := new(bytes.Buffer)

	// Encode image into PNG and write it to the buffer
	err = png.Encode(buf, codeImg)
	if err != nil {
		fmt.Println("error encode img: ", err)
	}

	// Convert the buffer to []bytes
	imageBytes := buf.Bytes()

	return imageBytes
}

func sendCell(bot *tgbotapi.BotAPI, chatID int64) error {
	switch {
	case line == 0:
		msg := tgbotapi.NewMessage(chatID, "Выберите ряд")

		keyboard := tgbotapi.ReplyKeyboardMarkup{
			InputFieldPlaceholder: "Выберите ряд",
			ResizeKeyboard: true,
		}

		for i := 1; i < 5; i++ {
			keyboard.Keyboard = append(keyboard.Keyboard, []tgbotapi.KeyboardButton{})
			for j := i*5 - 5; j < i*5; j++ {
				if j == 0 {
					continue
				}
				keyboard.Keyboard[i-1] = append(keyboard.Keyboard[i-1], tgbotapi.NewKeyboardButton(fmt.Sprintf("Ряд: "+strconv.Itoa(j))))
			}
		}

		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			fmt.Printf("Error sending photo: %s", err)
		}
	case cell == 0:
		msg := tgbotapi.NewMessage(chatID, "Выберите ячейку")

		keyboard := tgbotapi.ReplyKeyboardMarkup{
			InputFieldPlaceholder: "Выберите ячейку",
			ResizeKeyboard: true,
		}

		for i := 1; i < 4; i++ {
			keyboard.Keyboard = append(keyboard.Keyboard, []tgbotapi.KeyboardButton{})
			for j := i*3 - 3; j < i*3; j++ {
				if j == 0 {
					continue
				}
				keyboard.Keyboard[i-1] = append(keyboard.Keyboard[i-1], tgbotapi.NewKeyboardButton(fmt.Sprintf("Ячейка: "+strconv.Itoa(j))))
			}
		}

		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			fmt.Printf("Error sending photo: %s", err)
		}
	case floor == 0:
		msg := tgbotapi.NewMessage(chatID, "Выберите ярус")

		keyboard := tgbotapi.ReplyKeyboardMarkup{
			InputFieldPlaceholder: "Выберите ярус",
			ResizeKeyboard: true,
		}

		for i := 1; i < 4; i++ {
			keyboard.Keyboard = append(keyboard.Keyboard, []tgbotapi.KeyboardButton{})
				keyboard.Keyboard[0] = append(keyboard.Keyboard[0], tgbotapi.NewKeyboardButton(fmt.Sprintf("Ярус: "+strconv.Itoa(i))))
		}

		msg.ReplyMarkup = keyboard
		if _, err := bot.Send(msg); err != nil {
			fmt.Printf("Error sending photo: %s", err)
		}
	default:
		lineStr := strconv.Itoa(line)
		cellStr := strconv.Itoa(cell)
		floorStr := strconv.Itoa(floor)

		if line < 10 {
			lineStr = "0"+lineStr
		}
		if cell < 10 {
			cellStr = "0"+cellStr
		}

		text := "010_814-" + lineStr + "-" + cellStr + "-" + floorStr
		if err := sendMessage(bot, chatID, text); err != nil {
			fmt.Println("error in sending reply:", err)
		}
		line = 0
		cell = 0
		floor = 0
	}
	fmt.Printf("Line is: %d", line)
	return nil
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, code string) error {
	// Create image from text
	imageBytes := createImg(code)
	bytes := tgbotapi.FileBytes{Name: "image.png", Bytes: imageBytes}

	// Preparing message with image
	msg := tgbotapi.NewPhoto(chatID, bytes)

	// Add keyboard with buttons
	msg.ReplyMarkup = addKeyboard()

	msg.Caption = code

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
			tgbotapi.NewKeyboardButton("010_814-Exit_Dost"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Discount_814"),
			tgbotapi.NewKeyboardButton("V-Discount_814"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("010_814-Exit_sklad"),
			tgbotapi.NewKeyboardButton("010_814-Exit_zal"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Выбрать ячейку"),
		),
	)
	return keyboard
}

func main() {
	// Create a new HTTP server
	// Start the server on port 3000
	http.ListenAndServe(":3000", http.HandlerFunc(Handler))
}
