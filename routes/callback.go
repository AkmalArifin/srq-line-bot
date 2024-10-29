package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"example.com/yahfaz/models"
	"example.com/yahfaz/utils"
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

	/** Handle Learning */
	case "learn":
		state[userID] = "learn"

		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Please enter the Quran pages you want to add to your memorization.",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent text reply.")

	/** Handle Review */
	case "review":
		state[userID] = "review"

		memorizations, err := dueMemorization(userID)
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
						Text: "Sorry, you don't have any cards to reviewed. To check when you have to review, type 𝙨𝙩𝙖𝙩𝙪𝙨",
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

	/** Handle Status */
	case "status":
		status, err := statusMemorization(userID)

		if err != nil {
			log.Println(err.Error())
			return
		}

		review, err := dueMemorization(userID)

		text := ""

		text += fmt.Sprintf("𝙄𝙣 𝘿𝙪𝙚: %d\n", len(review))

		if err != nil {
			log.Println(err.Error())
			return
		}

		var timeKeys []time.Time
		for k := range status {
			timeKeys = append(timeKeys, k)
		}

		sort.Slice(timeKeys, func(i, j int) bool {
			return timeKeys[i].Before(timeKeys[j])
		})

		for i, timeKey := range timeKeys {
			if i == 0 {
				text += "Today:\n"
			} else {
				text += fmt.Sprintf("%s:\n", timeKey.Weekday().String())
			}
			var hourKeys []int
			for hourKey := range status[timeKey] {
				hourKeys = append(hourKeys, hourKey)
			}

			sort.Ints(hourKeys)

			for _, hourKey := range hourKeys {
				var noon string
				var time int
				if hourKey >= 12 {
					noon = "pm"
					time = hourKey - 12
				} else {
					noon = "am"
					time = hourKey
				}
				if time == 0 {
					time += 12
				}
				text += fmt.Sprintf(" - %d %s: %d\n", time, noon, status[timeKey][hourKey])

				// text += fmt.Sprintf(" - %d: %d\n", hourKey, status[timeKey][hourKey])
			}
			text += "\n"
		}

		_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: text,
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent status reply")
		return

	/** Handle Show */
	case "show":
		juzPages, err := showMemorizationPage(userID)

		if err != nil {
			log.Println(err.Error())
			return
		}

		var keys []int
		for k := range juzPages {
			keys = append(keys, k)
		}

		sort.Ints(keys)

		text := ""
		for _, k := range keys {
			text += fmt.Sprintf("Juz %d: ", k)
			for _, i := range juzPages[k] {
				text += fmt.Sprintf("%d ", i)
			}
			text += "\n"
		}

		if text == "" {
			text = "Sorry you don't have any memorization. To add pages to your memorization list, use 𝙡𝙚𝙖𝙧𝙣 command"
		}

		_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: text,
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent show text")

	/** Handle Help */
	case "help":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: `Yahfaz is a bot that supports you in memorizing the Quran using a spaced repetition system, helping you stay consistent even with a busy schedule.

𝙈𝙖𝙞𝙣 𝙁𝙚𝙖𝙩𝙪𝙧𝙚𝙨

1. Learn
Use the 𝙡𝙚𝙖𝙧𝙣 feature to log Quran pages you've memorized. Yahfaz accepts entries one page at a time.

2. Review
Yahfaz will schedule reviews for you based on spaced repetition principles, reminding you when it's time to revisit a page. To see which pages are scheduled for review, use the 𝙨𝙩𝙖𝙩𝙪𝙨 command. When you're ready, use the 𝙧𝙚𝙫𝙞𝙚𝙬 command and assess your memorization for each page:
	• Easy: 0-2 mistakes (review interval increases).
	• Good: 3-4 mistakes.
	• Hard: 5+ mistakes (review interval shortens).

With Yahfaz, you can keep track of your progress and review efficiently, ensuring long-term retention.
					`,
				},
				messaging_api.TextMessage{
					Text: `𝘼𝙫𝙖𝙞𝙡𝙖𝙗𝙡𝙚 𝘾𝙤𝙢𝙢𝙖𝙣𝙙𝙨:

𝙝𝙚𝙡𝙥              Help about guide and commands.
𝙝𝙚𝙡𝙥 𝙡𝙚𝙖𝙧𝙣      Help about learn command.
𝙝𝙚𝙡𝙥 𝙧𝙚𝙫𝙞𝙚𝙬    Help aabout review command.
𝙡𝙚𝙖𝙧𝙣              Add page to your memorization list
𝙧𝙚𝙫𝙞𝙚𝙬           Reviewing page based on spaced repetition system
𝙨𝙝𝙤𝙬             Show your memorization list
𝙨𝙩𝙖𝙩𝙪𝙨            Show review forecast
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

	case "help learn":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: `𝙇𝙚𝙖𝙧𝙣 𝘾𝙤𝙢𝙢𝙖𝙣𝙙

Use this command to add pages to your memorization list. After you've memorized a Quran page, add it here. Yahfaz will notify you when it's time to review this page using the 𝙧𝙚𝙫𝙞𝙚𝙬 command.`,
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent help text")
		return

	case "help review":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: `𝙍𝙚𝙫𝙞𝙚𝙬 𝘾𝙤𝙢𝙢𝙖𝙣𝙙

This command shows which pages in your memorization list are due for review. You won’t review everything at once; each page reappears based on your previous responses, with 𝙚𝙖𝙨𝙮 reviews taking longer to reappear.

Review Guidelines:
	• Easy: 0-2 mistakes
	• Good: 3-5 mistakes
	• Hard: 5+ mistakes

Answer as honestly as possible for effective scheduling. For accurate reviews, consider asking a friend to listen or using another memorization app.

𝗖𝗼𝗻𝘀𝗶𝘀𝘁𝗲𝗻𝗰𝘆 𝗶𝘀 𝗸𝗲𝘆: Try to review daily. Although it may feel slow, remember that the Prophet Muhammad (peace be upon him) said, "𝘛𝘩𝘦 𝘮𝘰𝘴𝘵 𝘣𝘦𝘭𝘰𝘷𝘦𝘥 𝘢𝘤𝘵𝘴 𝘰𝘧 𝘸𝘰𝘳𝘴𝘩𝘪𝘱 𝘢𝘳𝘦 𝘵𝘩𝘰𝘴𝘦 𝘵𝘩𝘢𝘵 𝘢𝘳𝘦 𝘤𝘰𝘯𝘴𝘪𝘴𝘵𝘦𝘯𝘵, 𝘦𝘷𝘦𝘯 𝘪𝘧 𝘵𝘩𝘦𝘺 𝘢𝘳𝘦 𝘴𝘮𝘢𝘭𝘭” (Sahih Muslim 783).
					`,
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent help text")
		return

	/** Handle Default */
	default:
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Please only input 𝙡𝙚𝙖𝙧𝙣, 𝙧𝙚𝙫𝙞𝙚𝙬, 𝙨𝙩𝙖𝙩𝙪𝙨, 𝙨𝙝𝙤𝙬, or 𝙝𝙚𝙡𝙥 if you want to know the details",
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

	/** Handle Canceling */
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

	/** Handle Help */
	case "help":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: `𝙇𝙚𝙖𝙧𝙣 𝘾𝙤𝙢𝙢𝙖𝙣𝙙

Use this command to add pages to your memorization list. After you've memorized a Quran page, add it here. Yahfaz will notify you when it's time to review this page using the 𝙧𝙚𝙫𝙞𝙚𝙬 command.`,
				},
				messaging_api.TextMessage{
					Text: "Please enter the Quran pages you want to add to your memorization.",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Sent help text")
		return

	/** Handle Input Learning */
	default:
		pageID, err := strconv.ParseInt(message.Text, 10, 64)

		if err != nil {
			_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
				ReplyToken: replyToken,
				Messages: []messaging_api.MessageInterface{
					messaging_api.TextMessage{
						Text: "Please enter number only, 𝙘𝙖𝙣𝙘𝙚𝙡 if you want to cancel, or 𝙝𝙚𝙡𝙥 if you want to know the details",
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

		// Input validation
		if pageID < 1 || pageID > 604 {
			_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
				ReplyToken: replyToken,
				Messages: []messaging_api.MessageInterface{
					messaging_api.TextMessage{
						Text: fmt.Sprintf("There is no page %d in Quran Mushaf Utsmani", pageID),
					},
					messaging_api.TextMessage{
						Text: "Please enter the Quran pages you want to add to your memorization.",
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

	/** Handle Yes */
	case "yes":
		err := createMemorizationPage(learnState[userID], userID)

		if err != nil {
			if !utils.IsDuplicateError(err.Error()) {
				log.Println(err.Error())
				return
			}

			_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
				ReplyToken: replyToken,
				Messages: []messaging_api.MessageInterface{
					messaging_api.TextMessage{
						Text: fmt.Sprintf("You've already added page %d", learnState[userID]),
					},
					messaging_api.TextMessage{
						Text: "Please enter the Quran pages you want to add to your memorization.",
					},
				},
			})

			if err != nil {
				log.Println(err.Error())
				return
			}

			state[userID] = "learn"
			log.Println("Sent error text")
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
		return

	/** Handle No */
	case "no":
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Please enter the Quran pages you want to add to your memorization.",
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

	default:
		_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: "Please enter 𝙮𝙚𝙨 or 𝙣𝙤.",
				},
			},
		})

		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println("Default text sent")
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

		memorizations, err := dueMemorization(userID)

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
						Text: "Congratulation, all your review pages are reviewed.",
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
		_, err = bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: fmt.Sprintf("Page %d reviewed.", reviewState[userID].page),
				},
				messaging_api.TextMessage{
					Text: fmt.Sprintf("Review Page %d", memorizations[0].PageID.Int64),
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

		reviewState[userID] = tuple{memorizations[0].ID, memorizations[0].PageID.Int64}

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
					Text: `𝙍𝙚𝙫𝙞𝙚𝙬 𝘾𝙤𝙢𝙢𝙖𝙣𝙙

This command shows which pages in your memorization list are due for review. You won’t review everything at once; each page reappears based on your previous responses, with 𝙚𝙖𝙨𝙮 reviews taking longer to reappear.

Review Guidelines:
	• Easy: 0-2 mistakes
	• Good: 3-5 mistakes
	• Hard: 5+ mistakes

Answer as honestly as possible for effective scheduling. For accurate reviews, consider asking a friend to listen or using another memorization app.

𝗖𝗼𝗻𝘀𝗶𝘀𝘁𝗲𝗻𝗰𝘆 𝗶𝘀 𝗸𝗲𝘆: Try to review daily. Although it may feel slow, remember that the Prophet Muhammad (peace be upon him) said, "𝘛𝘩𝘦 𝘮𝘰𝘴𝘵 𝘣𝘦𝘭𝘰𝘷𝘦𝘥 𝘢𝘤𝘵𝘴 𝘰𝘧 𝘸𝘰𝘳𝘴𝘩𝘪𝘱 𝘢𝘳𝘦 𝘵𝘩𝘰𝘴𝘦 𝘵𝘩𝘢𝘵 𝘢𝘳𝘦 𝘤𝘰𝘯𝘴𝘪𝘴𝘵𝘦𝘯𝘵, 𝘦𝘷𝘦𝘯 𝘪𝘧 𝘵𝘩𝘦𝘺 𝘢𝘳𝘦 𝘴𝘮𝘢𝘭𝘭” (Sahih Muslim 783).
					`,
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
					Text: "Please only input the 𝙚𝙖𝙨𝙮, 𝙜𝙤𝙤𝙙, 𝙝𝙖𝙧𝙙, 𝙘𝙖𝙣𝙘𝙚𝙡 if you want to cancel, or 𝙝𝙚𝙡𝙥 if you want to know the details",
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
