package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

// type rawFlexContainer struct {
// 	Type      FlexContainerType `json:"type"`
// 	Container FlexContainer     `json:"-"`
// }

func main() {
	err0 := godotenv.Load(fmt.Sprintf("%s.env", os.Getenv("GO_ENV")))
	if err0 != nil {
		fmt.Println("could not load env file")
	}
	// LINE Botクライアント生成する
	// BOT にはチャネルシークレットとチャネルトークンを環境変数から読み込み引数に渡す
	bot, err := linebot.New(
		os.Getenv("SECRET"),
		os.Getenv("CHANNEL_ACCESS_TOKEN"),
	)
	// エラーに値があればログに出力し終了する
	if err != nil {
		log.Fatal(err)
	}
	// テキストメッセージを生成する
	message := linebot.NewTextMessage("hello, world")
	// テキストメッセージを友達登録しているユーザー全員に配信する
	if _, err := bot.BroadcastMessage(message).Do(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("send")
	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				case *linebot.StickerMessage:
					replyMessage := fmt.Sprintf(
						"sticker id is %s, stickerResourceType is %s", message.StickerID, message.StickerResourceType)
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}

	// jsonData := []byte(`
	// {
	// 	"type": "bubble",
	// 	"hero": {
	// 	  "type": "box",
	// 	  "layout": "vertical",
	// 	  "contents": [
	// 		{
	// 		  "type": "text",
	// 		  "text": "site",
	// 		  "size": "xl",
	// 		  "color": "#ffffff",
	// 		  "align": "center"
	// 		}
	// 	  ],
	// 	  "backgroundColor": "#666666"
	// 	},
	// 	"body": {
	// 	  "type": "box",
	// 	  "layout": "vertical",
	// 	  "contents": [
	// 		{
	// 		  "type": "text",
	// 		  "text": "tsuntsunでサイトを積み始めたら爆速で消化できるようになった話",
	// 		  "weight": "bold",
	// 		  "size": "xl",
	// 		  "wrap": true
	// 		},
	// 		{
	// 		  "type": "box",
	// 		  "layout": "vertical",
	// 		  "margin": "lg",
	// 		  "spacing": "sm",
	// 		  "contents": [
	// 			{
	// 			  "type": "box",
	// 			  "layout": "baseline",
	// 			  "spacing": "sm",
	// 			  "contents": [
	// 				{
	// 				  "type": "text",
	// 				  "text": "URL",
	// 				  "color": "#aaaaaa",
	// 				  "size": "sm",
	// 				  "flex": 2
	// 				},
	// 				{
	// 				  "type": "text",
	// 				  "text": "http://localhost:8080",
	// 				  "wrap": true,
	// 				  "color": "#666666",
	// 				  "size": "sm",
	// 				  "flex": 5
	// 				}
	// 			  ]
	// 			},
	// 			{
	// 			  "type": "box",
	// 			  "layout": "baseline",
	// 			  "spacing": "sm",
	// 			  "contents": [
	// 				{
	// 				  "type": "text",
	// 				  "text": "created",
	// 				  "color": "#aaaaaa",
	// 				  "size": "sm",
	// 				  "flex": 2,
	// 				  "wrap": true
	// 				},
	// 				{
	// 				  "type": "text",
	// 				  "text": "2021/07/02",
	// 				  "wrap": true,
	// 				  "color": "#666666",
	// 				  "size": "sm",
	// 				  "flex": 5
	// 				}
	// 			  ]
	// 			},
	// 			{
	// 			  "type": "box",
	// 			  "layout": "baseline",
	// 			  "spacing": "sm",
	// 			  "contents": [
	// 				{
	// 				  "type": "text",
	// 				  "text": "total time",
	// 				  "color": "#aaaaaa",
	// 				  "size": "sm",
	// 				  "flex": 2,
	// 				  "wrap": true
	// 				},
	// 				{
	// 				  "type": "text",
	// 				  "text": "5min",
	// 				  "wrap": true,
	// 				  "color": "#666666",
	// 				  "size": "sm",
	// 				  "flex": 5
	// 				}
	// 			  ]
	// 			}
	// 		  ]
	// 		}
	// 	  ]
	// 	},
	// 	"footer": {
	// 	  "type": "box",
	// 	  "layout": "vertical",
	// 	  "spacing": "sm",
	// 	  "contents": [
	// 		{
	// 		  "type": "button",
	// 		  "style": "link",
	// 		  "height": "sm",
	// 		  "action": {
	// 			"type": "uri",
	// 			"label": "read now",
	// 			"uri": "https://linecorp.com"
	// 		  }
	// 		},
	// 		{
	// 		  "type": "button",
	// 		  "style": "link",
	// 		  "height": "sm",
	// 		  "action": {
	// 			"type": "uri",
	// 			"label": "already read",
	// 			"uri": "https://linecorp.com"
	// 		  }
	// 		},
	// 		{
	// 		  "type": "spacer",
	// 		  "size": "sm"
	// 		}
	// 	  ],
	// 	  "flex": 0
	// 	}
	//   }
	// `)
	// container, err := linebot.UnmarshalFlexMessageJSON(jsonData)
	// if err != nil {
	// 	// 正しくUnmarshalできないinvalidなJSONであればerrが返る
	// 	fmt.Println("could not form json data")
	// }
	// message1 := linebot.NewFlexMessage("alt text", container)
}

// func UnmarshalFlexMessageJSON(data []byte) (FlexContainer, error) {
// 	raw := rawFlexContainer{}
// 	if err := json.Unmarshal(data, &raw); err != nil {
// 		return nil, err
// 	}
// 	return raw.Container, nil
// }

// func (c *rawFlexContainer) UnmarshalJSON(data []byte) error {
// 	type alias rawFlexContainer
// 	raw := alias{}
// 	if err := json.Unmarshal(data, &raw); err != nil {
// 		return err
// 	}
// 	var container FlexContainer
// 	switch raw.Type {
// 	case FlexContainerTypeBubble:
// 		container = &BubbleContainer{}
// 	case FlexContainerTypeCarousel:
// 		container = &CarouselContainer{}
// 	default:
// 		return errors.New("invalid container type")
// 	}
// 	if err := json.Unmarshal(data, container); err != nil {
// 		return err
// 	}
// 	c.Type = raw.Type
// 	c.Container = container
// 	return nil
// }
