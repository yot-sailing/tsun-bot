package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot"
)

type Tsundokus struct {
	Title    string
	Category int // 0 => book, 1 => site
	URL      string
	// Author string
	RequiredTime string
	createdAt    string
}

func main() {
	want_added := false
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
					if message.Text == "今暇" {
						want_added = false
						resp := linebot.NewTemplateMessage(
							"this is a buttons template",
							linebot.NewButtonsTemplate(
								"./tsn.jpg",
								"積ん読消化！！",
								"何時間何分暇か選んでね",
								linebot.NewDatetimePickerAction("Time", "datetimepicker", "time", "", "23:59", "00:00"),
							),
						)

						_, err = bot.ReplyMessage(event.ReplyToken, resp).Do()
						if err != nil {
							log.Print(err)
						}
					} else if message.Text == "積みます" {
						resp := linebot.NewTemplateMessage(
							"this is a confirm template",
							linebot.NewConfirmTemplate(
								"本を積みますか?サイトを積みますか??",
								linebot.NewMessageAction("本", "本"),
								linebot.NewMessageAction("サイト", "サイト"),
							),
						)

						_, err = bot.ReplyMessage(event.ReplyToken, resp).Do()
						if err != nil {
							log.Print(err)
						}
					} else if message.Text == "今の積ん読リストを見せて" {
						want_added = false
						results := []Tsundokus{}
						//ここでAPIを呼び出す
						jsonData := (`
					{
					"type": "carousel",
					"contents": [
					`)
						for i, a := range results {
							jsonData += (`
								{
								"type": "bubble",
								"body": {
									"type": "box",
									"layout": "vertical",
									"contents": [
									{
										"type": "text",
										"text": 
									`) + a.Title
							// "tsuntsunでサイトを積み始めたら爆速で消化できるようになった話"
							jsonData += (`
										,
										"weight": "bold",
										"size": "xl",
										"wrap": true
									},
									{
										"type": "box",
										"layout": "vertical",
										"margin": "lg",
										"spacing": "sm",
										"contents": [
										{
											"type": "box",
											"layout": "baseline",
											"spacing": "sm",
											"contents": [
											{
												"type": "text",
												"text": "URL",
												"color": "#aaaaaa",
												"size": "sm",
												"flex": 2
											},
											{
											  "type": "text",
											  "text":`) + a.URL
							//   "http://localhost:8080"
							jsonData += (`
											  ,
											  "wrap": true,
											  "color": "#666666",
											  "size": "sm",
											  "flex": 5
											}
										  ]
										},
										{
										  "type": "box",
										  "layout": "baseline",
										  "spacing": "sm",
										  "contents": [
											{
											  "type": "text",
											  "text": "created",
											  "color": "#aaaaaa",
											  "size": "sm",
											  "flex": 2,
											  "wrap": true
											},
											{
											  "type": "text",
											  `) + a.createdAt
							//   "text": "2021/07/02"
							jsonData += (`
											  ,
											  "wrap": true,
											  "color": "#666666",
											  "size": "sm",
											  "flex": 5
											}
										  ]
										},
										{
										  "type": "box",
										  "layout": "baseline",
										  "spacing": "sm",
										  "contents": [
											{
											  "type": "text",
											  "text": "total time",
											  "color": "#aaaaaa",
											  "size": "sm",
											  "flex": 2,
											  "wrap": true
											},
											{
											  "type": "text",
											  "text": 
											  `) + a.RequiredTime
							//   "5min"
							jsonData += (`
											  ,
											  "wrap": true,
											  "color": "#666666",
											  "size": "sm",
											  "flex": 5
											}
										  ]
										}
									  ]
									}
								  ]
								},
								"footer": {
								  "type": "box",
								  "layout": "vertical",
								  "spacing": "sm",
								  "contents": [
									{
									  "type": "button",
									  "style": "link",
									  "height": "sm",
									  "action": {
										"type": "uri",
										"label": "read now",
										"uri": "https://linecorp.com"
									  }
									},
									{
									  "type": "button",
									  "style": "link",
									  "height": "sm",
									  "action": {
										"type": "uri",
										"label": "already read",
										"uri": "https://linecorp.com"
									  }
									},
									{
									  "type": "spacer",
									  "size": "sm"
									}
								  ],
								  "flex": 0
								}
							  }
							`)
							if i != len(results)-1 {
								jsonData += ","
							}
						}
						// fmt.Println(jsonData)
						jsonData += "]}"
						fmt.Println(jsonData)
						container, err_f := linebot.UnmarshalFlexMessageJSON([]byte(jsonData))
						if err_f != nil {
							fmt.Println("could not read json data because of ", err_f)
						}
						if _, err4 := bot.ReplyMessage(
							event.ReplyToken,
							linebot.NewFlexMessage("tsuntsun-list", container),
						).Do(); err4 != nil {
							fmt.Println(err4)
						}
					} else if message.Text == "サイト" {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("URLちょうだい！")).Do(); err != nil {
							log.Print(err)
						}
						want_added = true
					} else if message.Text == "本" {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("タイトルを教えて")).Do(); err != nil {
							log.Print(err)
						}
						want_added = true
						//ここで 積ん読追加のAPIを呼ぶ、著者とタイトル、どう判断すべきか分からんからタイトルだけで
					} else if strings.Contains(message.Text, "http") {
						url := message.Text
						fmt.Println(url)
						//ここで 積んサイト追加のAPIを呼ぶ
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("追加したよ、はよ消化してね")).Do(); err != nil {
							log.Print(err)
						}
					} else {
						if want_added {
							title := message.Text
							fmt.Println(title)
							want_added = false
							//ここで 積ん読追加のAPIを呼ぶ
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("追加したよ、はよ消化してね")).Do(); err != nil {
								log.Print(err)
							}
						} else {
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
								log.Print(err)
							}
							want_added = false // ほんま？？
						}
					}
				case *linebot.StickerMessage:
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("スタンプよりも積ん読消化して")).Do(); err != nil {
						log.Print(err)
					}

				}
			} else if event.Type == linebot.EventTypePostback {
				fmt.Println(event.Postback.Params)
				//ここで何分で読めるサイトかを提案するAPIを呼び出す
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(event.Postback.Params.Time[:2]+"時間"+event.Postback.Params.Time[3:]+"分暇なのね")).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	})
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
