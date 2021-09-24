package util

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"main/model"
	"net/http"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/lib/pq"
	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

func GetTsundokus(DB *sql.DB, userID int) ([]model.Tsundoku, error) {
	var results []model.Tsundoku
	rows, err := DB.Query("select * from tsundokus where user_id = $1;", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var result model.Tsundoku
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

func CountRequiredTime(url string) (int, string, error) {
	res, err := http.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, "", nil
	}

	det := chardet.NewTextDetector()
	detResult, err := det.DetectBest(buf)

	if err != nil {
		return 0, "", err
	}

	bReader := bytes.NewReader(buf)
	reader, err := charset.NewReaderLabel(detResult.Charset, bReader)

	if err != nil {
		return 0, "", nil
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return 0, "", nil
	}
	doc.Find("*:empty").Remove()
	doc.Find("script").Remove()
	doc.Find("style").Remove()
	totalContents := utf8.RuneCountInString(doc.Find("body").Text())
	return totalContents / 500, doc.Find("title").Text(), nil
}
