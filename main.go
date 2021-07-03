package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/yukihir0/gec"
)

type Tsundoku struct {
	ID           int
	UserID       int
	Category     string
	Title        string
	Author       string
	URL          string
	Deadline     time.Time
	RequiredTime string
	CreatedAt    time.Time
}
type Book struct {
	Title  string
	Author string
}

var tsun_book Book

func main() {
	want_added := false  //本を加えたそう
	title_added := false //タイトルを加えてもらっています
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
				fmt.Println(event.Source.UserID)
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					if message.Text == "今暇" {
						want_added = false
						resp := linebot.NewTemplateMessage(
							"this is a buttons template",
							linebot.NewButtonsTemplate(
								"https://ddnavi.com/wp-content/uploads/2020/04/tsundoku.jpg",
								"積ん読消化！！",
								"何時間何分暇か選んでね",
								linebot.NewDatetimePickerAction("Time", "time", "time", "", "23:59", "00:00"),
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
						var results []Tsundoku
						if resp, err := http.Get("https://tsuntsun-api.herokuapp.com/api/users/1/tsundokus"); err != nil {
							fmt.Println("error:http get\n", err)
						} else {
							defer resp.Body.Close()
							byteArray, _ := ioutil.ReadAll(resp.Body)
							res_str := string(byteArray)
							err := json.Unmarshal([]byte(res_str), &results)
							if err != nil {
								fmt.Println(err)
								return
							}
						}
						fmt.Println(results)
						jsonData := (`
									{
									"type": "carousel",
									"contents": [`)
						for i, a := range results {
							column1 := ""
							column2 := ""
							image_url := "https://pakutaso.cdn.rabify.me/shared/img/thumb/macbookFTHG1289.jpg?d=350" // pc用
							if a.Category == "book" {                                                                // if book
								column1 = "author"
								column2 = "deadline"
								if a.Author == "" {
									a.URL = "まだ入力されてないヨ"
								} else {
									a.URL = a.Author //ここちょっと汚い
								}
								image_url = "https://imgs.u-note.me/note/caption/47488447.jpg"
								a.RequiredTime = a.Deadline.String()[:10]
							} else { // if site
								column1 = "URL"
								column2 = "total time"
							}
							jsonData += (`
								{
								"type": "bubble",
								"hero": {
									"type": "image",
									"url": "` + image_url + `",
									"size": "full",
									"aspectRatio": "20:13",
									"aspectMode": "cover"
								},
								"body": {
									"type": "box",
									"layout": "vertical",
									"contents": [
									{
										"type": "text",
										"text": "` + a.Title + `",
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
												"text": "` + column1 + `",
												"color": "#aaaaaa",
												"size": "sm",
												"flex": 2
											},
											{
											  "type": "text",
											  "text": "` + a.URL + `",
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
											  "text" : "` + a.CreatedAt.String()[:10] + `", 
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
											  "text": "` + column2 + `",
											  "color": "#aaaaaa",
											  "size": "sm",
											  "flex": 2,
											  "wrap": true
											},
											{
											  "type": "text",
											  "text": "` + a.RequiredTime + `" ,
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
								},`)
							if a.Category == "site" {
								jsonData += (`
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
										"uri": "` + a.URL + `"
									}
									},
									{
										"type": "button",
										"style": "link",
										"height": "sm",
										"action": {
											"type": "message",
											"label": "already read",
											"text": "already read : tsundokuID ` + strconv.Itoa(a.ID) + `"
										}
										},
									{
										"type": "spacer",
										"size": "sm"
									}
									],
									"flex": 0
									}
								}`)
							} else {
								jsonData += (`
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
										"type": "message",
										"label": "already read",
										"text": "already read : tsundokuID ` + strconv.Itoa(a.ID) + `"
									}
								},
								{
									"type": "spacer",
									"size": "sm"
								}
							],
							"flex": 0
							}
							}`)
							}
							if i != len(results)-1 {
								jsonData += ","
							}
						}
						jsonData += "]}"
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
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("タイトルを教えて、著者もわかるなら改行して入力して")).Do(); err != nil {
							log.Print(err)
						}
						want_added = true
						//ここで 積ん読追加のAPIを呼ぶ、著者とタイトル、どう判断すべきか分からんからタイトルだけで
					} else if strings.Contains(message.Text, "http") {
						tsumu_url := message.Text
						fmt.Println(tsumu_url)
						// URLからタイトルと本文の長さを取得する
						doc, err := goquery.NewDocument(tsumu_url)
						if err != nil {
							fmt.Println("err", err)
						}
						html, err := doc.Html()
						if err != nil {
							fmt.Println("err", err)
						}
						opt := gec.NewOption()
						content, title := gec.Analyse(html, opt)
						args := url.Values{}
						args.Add("category", "site")
						args.Add("url", tsumu_url)
						args.Add("title", title)
						args.Add("requiredTime", strconv.Itoa(len(content)/500)+"min")
						fmt.Println(strconv.Itoa(len(content)/500) + "min")
						_, err = http.PostForm("https://tsuntsun-api.herokuapp.com/api/users/1/tsundokus", args)
						if err != nil {
							fmt.Println("Request error:", err)
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("追加できなかったわ、ごめん")).Do(); err != nil {
								log.Print(err)
							}
						}
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("追加したよ、はよ消化してね")).Do(); err != nil {
							log.Print(err)
						}
					} else if strings.Contains(message.Text, "already read : tsundokuID ") {
						tsum_del, _ := strconv.Atoi(message.Text[26:])
						fmt.Println(tsum_del)
						req, _ := http.NewRequest("DELETE", "https://tsuntsun-api.herokuapp.com/api/users/1/tsundokus/"+strconv.Itoa(tsum_del), nil)
						req.Header.Set("Accept", "application/json")
						client := new(http.Client)
						_, err := client.Do(req)
						if err != nil {
							fmt.Println("Request error:", err)
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("消せなかった、すまぬ")).Do(); err != nil {
								log.Print(err)
							}
						}
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("えらい！よく読めました！")).Do(); err != nil {
							log.Print(err)
						}
					} else {
						if want_added { //本の追加
							title_author := message.Text
							if strings.Contains(title_author, "\n") {
								re := strings.Split(title_author, "\n")
								title := re[0]
								author := re[1]
								tsun_book.Author = author
								tsun_book.Title = title
							} else {
								title := title_author
								tsun_book.Title = title
							}
							want_added = false
							title_added = true
							resp := linebot.NewTemplateMessage(
								"this is a buttons template",
								linebot.NewButtonsTemplate(
									"https://ddnavi.com/wp-content/uploads/2020/04/tsundoku.jpg",
									"本をいつまでに読むか決めます",
									"何月何日に読み終えたいか教えてね",
									linebot.NewDatetimePickerAction("Date", "date", "date", "", "2025-07-02", "2021-07-02"),
								),
							)
							_, err = bot.ReplyMessage(event.ReplyToken, resp).Do()
							if err != nil {
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
				fmt.Println(event.Postback)
				if event.Postback.Data == "time" {
					//ここで何分で読めるサイトかを提案するAPIを呼び出す
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(event.Postback.Params.Time[:2]+"時間"+event.Postback.Params.Time[3:]+"分暇なのね")).Do(); err != nil {
						log.Print(err)
					}
				} else if event.Postback.Data == "date" && title_added {
					args := url.Values{}
					args.Add("category", "book")
					args.Add("title", tsun_book.Title)
					if tsun_book.Author != "" {
						args.Add("author", tsun_book.Author)
					}
					_, err = http.PostForm("https://tsuntsun-api.herokuapp.com/api/users/1/tsundokus", args)
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("追加したよ、はよ消化してね")).Do(); err != nil {
						log.Print(err)
					}
					title_added = false
					tsun_book = Book{}
				}

			}
		}
	})
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
