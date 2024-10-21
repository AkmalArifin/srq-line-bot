package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"example.com/yahfaz/models"
	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type tuple struct {
	id   int64
	page int64
}

var state = make(map[string]string)
var reviewState = make(map[string]tuple)
var learnState = make(map[string]int64)

func callbackHandler(c *gin.Context) {
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")

	bot, err := messaging_api.NewMessagingApiAPI(os.Getenv("TOKEN"))

	if err != nil {
		log.Println(err)
		return
	}

	cb, err := webhook.ParseRequest(channelSecret, c.Request)

	if err != nil {
		log.Println("cannot parse request")
		return
	}

	log.Println("Handle Request")
	for _, event := range cb.Events {
		// log.Printf("/callback called%+v...\n", event)

		switch e := event.(type) {
		case webhook.FollowEvent:
			// Get Data
			var userID string
			switch s := e.Source.(type) {
			case webhook.UserSource:
				userID = s.UserId
			}

			followHandler(userID)
		case webhook.MessageEvent:
			switch message := e.Message.(type) {
			case webhook.TextMessageContent:
				// Get Data
				var userID string
				switch s := e.Source.(type) {
				case webhook.UserSource:
					userID = s.UserId
				}

				userState, ok := state[userID]

				if !ok {
					userState = "idle"
				}

				switch userState {
				case "idle":
					idleStateHandler(bot, e.ReplyToken, message, userID)
				case "learn":
					learnStateHandler(bot, e.ReplyToken, message, userID)
				case "confirm":
					confirmStateHandler(bot, e.ReplyToken, message, userID)
				case "review":
					reviewStateHandler(bot, e.ReplyToken, message, userID)
				default:
					log.Printf("Unsupported state: %s\n", userState)
				}
			}
		}
	}
}

func idleStateHandler(bot *messaging_api.MessagingApiAPI, replyToken string, message webhook.TextMessageContent, userID string) {
	switch strings.ToLower(message.Text) {
	case "learn":
		state[userID] = "learn"

		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Please input quran pages that you want to add into your memorization",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent text reply.")

	case "review":
		state[userID] = "review"

		memorizations, err := models.GetReviewByUserID(userID)
		log.Println(memorizations)

		if err != nil {
			log.Println(err.Error())
			return
		}

		if len(memorizations) == 0 {
			_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
				ReplyToken: replyToken,
				Messages: []messaging_api.MessageInterface{
					messaging_api.TextMessage{
						Text: "Sorry, you don't have any cards to reviewed. To check when you have to review, type 'Status'",
					},
				},
			})

			if err != nil {
				log.Println(err.Error())
				return
			}

			log.Println("Sent reply text.")
			state[userID] = "idle"
			return
		}

		reviewState[userID] = tuple{memorizations[0].ID, memorizations[0].PageID.ValueOrZero()}

		_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: fmt.Sprintf("Review Page %d", reviewState[userID].page),
				},
				messaging_api.TemplateMessage{
					AltText: "Button template",
					Template: &messaging_api.ButtonsTemplate{
						Text: "How's your review?",
						Actions: []messaging_api.ActionInterface{
							messaging_api.MessageAction{
								Label: "Easy",
								Text:  "Easy",
							},
							messaging_api.MessageAction{
								Label: "Good",
								Text:  "Good",
							},
							messaging_api.MessageAction{
								Label: "Hard",
								Text:  "Hard",
							},
						},
					},
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent reply text")
		return

	case "status":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "In progress",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		status, err := statusMemorization(userID)

		if err != nil {
			log.Println(err.Error())
			return
		}

		for ind, stat := range status {
			log.Println(ind, stat)
		}

		log.Println("Sent text reply")
		return

	case "help":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Yahfaz is a simple bot that helped you to remind which pages of Quran that you need to review. It is using spaced repetition system in review system. It is made in order to accompany you for memorizing Quran while busy with works or studies and not having a full time dedication for memorizing Quran.",
				},
				messaging_api.TextMessage{
					// Text: fmt.Sprintf("Available Commands:\n  help\t\t\t\tHelp about any command\n  help\t\t\t\tAdd page to your memorization list"),
					Text: `Available Commands:
					help          Help about any command
					learn         Add page to your memorization list
					review       Reviewing page based on spaced repetition system
					status       Show review forecast
					`,
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Send help text")
		return

	default:
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Please only input 'Learn', 'Review', 'Status', or 'Help' if you want to know the details",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent text reply")
		return
	}
}

func learnStateHandler(bot *messaging_api.MessagingApiAPI, replyToken string, message webhook.TextMessageContent, userID string) {
	switch strings.ToLower(message.Text) {
	case "cancel":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Learning canceled",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		state[userID] = "idle"
		learnState[userID] = 0
		log.Println("Learning canceled")
		return

	case "help":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "This command is for adding page into your memorization list. After you memorized page of Quran, add the page into this command. Yahfaz will let you know, when you need to review this page later by using 'review' command.",
				},
				messaging_api.TextMessage{
					Text: "Please input quran pages that you want to add into your memorization",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent help text")
		return

	default:
		pageID, err := strconv.ParseInt(message.Text, 10, 64)

		if err != nil {
			_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
				ReplyToken: replyToken,
				Messages: []messaging_api.MessageInterface{
					messaging_api.TextMessage{
						Text: "Please only input the number, 'Cancel' if you want to cancel, or 'Help' if you want to know the details",
					},
				},
			})

			if err != nil {
				log.Println(err.Error())
				return
			}

			log.Println("Sent text reply")
			return
		}

		_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TemplateMessage{
					AltText: "Confirm template",
					Template: &messaging_api.ConfirmTemplate{
						Text: fmt.Sprintf("Are you sure you want to add page %d?", pageID),
						Actions: []messaging_api.ActionInterface{
							messaging_api.MessageAction{
								Label: "Yes",
								Text:  "Yes",
							},
							messaging_api.MessageAction{
								Label: "No",
								Text:  "No",
							},
						},
					},
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		state[userID] = "confirm"
		learnState[userID] = pageID
		log.Println("Confirm reply text")
		return
	}
}

func confirmStateHandler(bot *messaging_api.MessagingApiAPI, replyToken string, message webhook.TextMessageContent, userID string) {
	switch strings.ToLower(message.Text) {
	case "yes":
		err := createMemorizationPage(learnState[userID], userID)

		if err != nil {
			log.Println(err.Error())
			return
		}

		_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: fmt.Sprintf("Page %d has been added.", learnState[userID]),
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		state[userID] = "idle"
		learnState[userID] = 0
		log.Println("Memorization created")
	case "no":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Please input quran pages that you want to add into your memorization",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		state[userID] = "learn"
		learnState[userID] = 0
		log.Println("Confirmation no")
	}
}

func reviewStateHandler(bot *messaging_api.MessagingApiAPI, replyToken string, message webhook.TextMessageContent, userID string) {
	switch strings.ToLower(message.Text) {
	case "easy", "good", "hard":
		err := reviewMemorization(reviewState[userID].id, message.Text)

		if err != nil {
			log.Println(err.Error())
			return
		}

		memorizations, err := models.GetReviewByUserID(userID)

		if err != nil {
			log.Println(err.Error())
			return
		}

		if len(memorizations) == 0 {
			_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
				ReplyToken: replyToken,
				Messages: []messaging_api.MessageInterface{
					messaging_api.TextMessage{
						Text: fmt.Sprintf("Page %d reviewed.", reviewState[userID].page),
					},
					messaging_api.TextMessage{
						Text: "Congratulation, all your review cards are reviewed.",
					},
				},
			})

			if err != nil {
				log.Println(err.Error())
				return
			}

			log.Println("Sent reply text.")
			state[userID] = "idle"
			return
		}

		reviewState[userID] = tuple{memorizations[0].ID, memorizations[0].PageID.Int64}

		_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: fmt.Sprintf("Page %d reviewed.", reviewState[userID].page),
				},
				messaging_api.TextMessage{
					Text: fmt.Sprintf("Review Page %d", reviewState[userID].page),
				},
				messaging_api.TemplateMessage{
					AltText: "Button template",
					Template: &messaging_api.ButtonsTemplate{
						Text: "How's your review?",
						Actions: []messaging_api.ActionInterface{
							messaging_api.MessageAction{
								Label: "Easy",
								Text:  "Easy",
							},
							messaging_api.MessageAction{
								Label: "Good",
								Text:  "Good",
							},
							messaging_api.MessageAction{
								Label: "Hard",
								Text:  "Hard",
							},
						},
					},
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent reply text")
		return

	case "cancel":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Reviewing canceled",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		state[userID] = "idle"
		reviewState[userID] = tuple{0, 0}
		log.Println("Reviewing canceled")
		return

	case "help":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "This command is to help you know which page from list of your memorization that you should review. You won't be able review all your memorization at once. Every time you answer 'easy' for your review, it will show in review longer than before. Please answer as honest as possible.",
				},
				messaging_api.TextMessage{
					Text: "For tips, you can ask for your friend to check on your memorization or you can use another apps that could help you. For our ground rules. Easy if you made 0-2 mistakes. Good if you made 3-5 mistakes. Hard if you made more than 5 mistakes.",
				},
				messaging_api.TextMessage{
					Text: "The key is to be consistent, every day at least you check once what are the pages that you need to review. Yes it will take much longer, but The Prophet Muhammad told us that the most beloved acts of worship are those that are consistent, even if they are small (Sahih Muslim 783).",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent help text")
		return

	default:
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Please only input the 'Easy', 'Good', 'Hard', 'Cancel' if you want to cancel, or 'Help' if you want to know the details",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}
		log.Println("Sent text reply")
		return
	}
}

func followHandler(userID string) {
	url := fmt.Sprintf("https://api.line.me/v2/bot/profile/%s", userID)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Println(err.Error())
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TOKEN")))

	client := &http.Client{}

	response, err := client.Do(req)

	if err != nil {
		log.Println(err.Error())
		return
	}

	defer response.Body.Close()

	var profile models.Profile
	err = json.NewDecoder(response.Body).Decode(&profile)

	if err != nil {
		log.Println(err.Error())
		return
	}

	var user models.User
	user.UserID.SetValid(profile.UserID.ValueOrZero())
	user.DisplayName.SetValid(profile.DisplayName.ValueOrZero())
	user.Language.SetValid(profile.Language.ValueOrZero())
	err = user.Save()

	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println("Profile saved")
}
