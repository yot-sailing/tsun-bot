package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/yukihir0/gec"
)

func main() {
	doc, err := goquery.NewDocument("https://ledge.ai/bert/")
	if err != nil {
		println("err", err)
	}
	html, err := doc.Html()
	if err != nil {
		println("err", err)
	}
	opt := gec.NewOption()
	content, title := gec.Analyse(html, opt)
	println(len(content)/500, title)

	// g := goose.New()
	// article, _ := g.ExtractFromURL("https://ledge.ai/bert/")
	// println("title", article.Title)
	// println("description", article.MetaDescription)
	// println("keywords", article.MetaKeywords)
	// println("content", article.CleanedText)
	// println("url", article.FinalURL)
	// println("top image", article.TopImage)
}
