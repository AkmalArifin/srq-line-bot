package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func echoBot(c *gin.Context) {
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")

	bot, err := messaging_api.NewMessagingApiAPI(os.Getenv("TOKEN"))

	if err != nil {
		log.Println(err)
		return
	}

	cb, err := webhook.ParseRequest(channelSecret, c.Request)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Cannot parse request"})
		return
	}

	log.Println("Handle request")
	for _, event := range cb.Events {
		log.Printf("/callback called%+v...\n", event)

		switch e := event.(type) {
		case webhook.MessageEvent:

			switch message := e.Message.(type) {
			case webhook.TextMessageContent:
				if _, err = bot.ReplyMessage(
					&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							messaging_api.TextMessage{
								Text: message.Text,
							},
						},
					},
				); err != nil {
					log.Print(err)
				} else {
					log.Println("Sent text reply")
				}
			case webhook.StickerMessageContent:
				replyMessage := fmt.Sprintf("sticker id is %s, stickerResourceType is %s", message.StickerId, message.StickerResourceType)
				if _, err = bot.ReplyMessage(
					&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							messaging_api.TextMessage{
								Text: replyMessage,
							},
						},
					},
				); err != nil {
					log.Print(err)
				} else {
					log.Println("Sent sticker reply.")
				}
			default:
				log.Printf("Unsupported message content: %T\n", e.Message)
			}
		default:
			log.Printf("Unsupported message: %T\n", event)
		}
	}
}
