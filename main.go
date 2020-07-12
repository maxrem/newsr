package main

import (
	"github.com/mmcdole/gofeed"
	"log"
	"time"
)

var googleNewsUrl = "https://news.google.com/news?hl=en-US&ceid=US:en&sort=date&gl=US&num=10&output=rss&q="

func main() {
	keyWord := "covid+89bio"

	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(googleNewsUrl+keyWord)

	if err != nil {
		log.Println(err)
	}

	log.Println(feed.Title)
	log.Println(feed.Len())
	for i := 0; i < feed.Len(); i++ {
		item := feed.Items[i]

		timeParsed, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", item.Published)
		if err != nil {
			log.Println(err)
		}

		timeSince := time.Since(timeParsed)

		if timeSince < 48 * time.Hour {
			log.Println(item.Title, timeParsed)
		}
	}
}
