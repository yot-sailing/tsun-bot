package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/PuerkitoBio/goquery"
	"github.com/lib/pq"
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

var DB *sql.DB

var tsun_book Book

func main() {
	var err error
	DB, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
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
				var userID int
				log.Println("70:line id is", event.Source.UserID)
				err := DB.QueryRow("select id from users where line_id = $1;", event.Source.UserID).Scan(&userID)
				if err != nil {
					log.Fatal(err)
					return
				}
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
								linebot.NewDatetimePickerAction("暇な時間入力", "time", "time", "00:00", "23:59", "00:00"),
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

						results, err := getTsundokus(userID)
						if err != nil {
							log.Println(err)
							return
						}

						if len(results) > 12 {
							results = results[:12]
						}
						if len(results) == 0 {
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("今積んでる本やサイトはないよ！")).Do(); err != nil {
								log.Print(err)
							}
							return
						}
						jsonData := (`
									{
									"type": "carousel",
									"contents": [`)
						for i, a := range results {
							column1 := ""
							column2 := ""
							image_url := "https://pakutaso.cdn.rabify.me/shared/img/thumb/macbookFTHG1289.jpg?d=350" // pc用
							if a.Category == "book" {                                                                // if book
								column1 = "著者"
								column2 = "この日までに読む"
								if a.Author == "" {
									a.URL = "まだ入力されてないヨ"
								} else {
									a.URL = a.Author //ここちょっと汚い
								}
								image_url = "https://imgs.u-note.me/note/caption/47488447.jpg"
								a.RequiredTime = a.Deadline.String()[:10]
							} else { // if site
								column1 = "URL"
								column2 = "読了に必要な時間"
								a.RequiredTime = a.RequiredTime + "分"
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
											  "text": "作成日時",
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
										"label": "今から読む",
										"uri": "` + a.URL + `"
									}
									},
									{
										"type": "button",
										"style": "link",
										"height": "sm",
										"action": {
											"type": "message",
											"label": "もう読んだよ",
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
										"label": "もう読んだよ",
										"text": "積ん読1つ消化！！(tsundokuID : ` + strconv.Itoa(a.ID) + `)"
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
						var tsundoku_id int
						time := time.Now()
						t := time.Format("2006-01-02")
						err = DB.QueryRow("INSERT INTO tsundokus (user_id, category, url, title, required_time, created_at) values ($1 , $2, $3, $4, $5, $6) RETURNING id;", userID, "site", tsumu_url, title, strconv.Itoa(len(content)/500), t).Scan(&tsundoku_id)
						if err != nil {
							log.Println(err)
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("追加できなかった、すまぬ")).Do(); err != nil {
								log.Print(err)
								return
							}
						}
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("追加したよ、はよ消化してね")).Do(); err != nil {
							log.Print(err)
						}
					} else if strings.Contains(message.Text, "積ん読1つ消化！！(tsundokuID ") {
						tsum_del, _ := strconv.Atoi(message.Text[26:])
						result, err := DB.Exec("DELETE FROM tsundokus WHERE id = $1;", strconv.Itoa(tsum_del)) //user_idを指定することでそのuserしか消せないようになるはず??
						if err != nil {
							log.Println(err)
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("消せなかった、すまぬ")).Do(); err != nil {
								log.Print(err)
							}
							return
						}
						_, err = result.RowsAffected()
						if err != nil {
							log.Println("Request error:", err)
							if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("消せなかった、すまぬ")).Do(); err != nil {
								log.Print(err)
							}
							return
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
									linebot.NewDatetimePickerAction("Date", "date", "date", "2021-09-18", "2025-07-02", "2021-09-18"),
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
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewStickerMessage("446", "1988")).Do(); err != nil {
						log.Print(err)
					}

				}
			} else if event.Type == linebot.EventTypePostback {
				var userID int
				err := DB.QueryRow("select id from users where line_id = $1;", event.Source.UserID).Scan(&userID)
				if err != nil {
					log.Fatal(err)
					return
				}
				if event.Postback.Data == "time" {
					limited_results := []Tsundoku{}
					hour, _ := strconv.Atoi(event.Postback.Params.Time[:2])
					min, _ := strconv.Atoi(event.Postback.Params.Time[3:])
					total_min := hour*60 + min
					results, err := getTsundokus(userID)
					if err != nil {
						log.Println(err)
						return
					}
					for _, element := range results {
						if element.Category == "site" {
							required_time, _ := strconv.Atoi(element.RequiredTime)
							fmt.Println("time api", required_time, total_min)
							if total_min >= required_time {
								limited_results = append(limited_results, element)
							}
						}
					}

					if len(limited_results) == 0 {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(strconv.Itoa(total_min)+"分以内で読めるサイトは無いわ、、")).Do(); err != nil {
							log.Print(err)
						}
						return
					}
					jsonData := (`
									{
									"type": "carousel",
									"contents": [`)
					for i, a := range limited_results {
						column1 := "URL"
						column2 := "total time"
						image_url := "https://pakutaso.cdn.rabify.me/shared/img/thumb/macbookFTHG1289.jpg?d=350" // pc用
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
										"text": "` + strconv.Itoa(total_min) + `分以内で読める"
									},
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
											  "weight": "bold",
											  "color": "#ef93b6",
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
										"label": "今から読む",
										"uri": "` + a.URL + `"
									}
									},
									{
										"type": "button",
										"style": "link",
										"height": "sm",
										"action": {
											"type": "message",
											"label": "もう読んだよ",
											"text": "積読1つ消化！！(tsundokuID : ` + strconv.Itoa(a.ID) + `)"
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
						if i != len(limited_results)-1 {
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
				} else if event.Postback.Data == "date" && title_added {
					var tsundoku_id int
					time := time.Now()
					t := time.Format("2006-01-02")
					err = DB.QueryRow("INSERT INTO tsundokus (user_id, category, title, author, deadline, created_at) values ($1 , $2, $3, $4, $5, $6);", userID, "book", tsun_book.Title, tsun_book.Author, event.Postback.Params.Date, t).Scan(&tsundoku_id)
					if err != nil {
						log.Println(err)
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("追加できなかった、すまぬ")).Do(); err != nil {
							log.Print(err)
							return
						}
					}
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

func getTsundokus(userID int) ([]Tsundoku, error) {
	var results []Tsundoku
	rows, err := DB.Query("select * from tsundokus where user_id = $1;", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var result Tsundoku
		nullAuthor := new(sql.NullString)
		nullURL := new(sql.NullString)
		nullDeadLine := new(pq.NullTime)
		nullRequiredTime := new(sql.NullString)
		nullCreatedAt := new(pq.NullTime)
		err := rows.Scan(&result.ID, &result.UserID, &result.Category, &result.Title, nullAuthor, nullURL, nullDeadLine, nullRequiredTime, nullCreatedAt)
		if err != nil {
			return nil, err
		}
		if nullAuthor.Valid {
			result.Author = nullAuthor.String
		}
		if nullURL.Valid {
			result.URL = nullURL.String
		}
		if nullDeadLine.Valid {
			result.Deadline = nullDeadLine.Time
		}
		if nullRequiredTime.Valid {
			result.RequiredTime = nullRequiredTime.String
		}
		if nullCreatedAt.Valid {
			result.CreatedAt = nullCreatedAt.Time
		}
		results = append(results, result)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
