package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"log"
	"net/url"
	"sync"
	"time"
)

const googleNewsUrl = "https://news.google.com/news?hl=en-US&ceid=US:en&sort=date&gl=US&num=10&output=rss&q="
const dateTimeLayout = "Mon, 02 Jan 2006 15:04:05 MST"
const keyWord = "covid"

// TODO https://stackoverflow.com/questions/55203251/limiting-number-of-go-routines-running

func main() {
	var wg sync.WaitGroup

	subjectList := []string{
		"89bio",
		"AC Immune SA",
		"ACADIA Pharmaceuticals Inc.",
		"Acasti Pharma, Inc.",
		"Accelerate Diagnostics, Inc.",
	}

	for _, subject := range subjectList {
		wg.Add(1)
		go parseFeed(googleNewsUrl+fmt.Sprintf("%s+%s", url.QueryEscape(keyWord), url.QueryEscape(subject)), &wg)
	}

	wg.Wait()
}

func parseFeed(url string, wg *sync.WaitGroup) {
	log.Println("Parsing", url)
	defer wg.Done()
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(url)

	if err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < feed.Len(); i++ {
		item := feed.Items[i]

		timeParsed, err := time.Parse(dateTimeLayout, item.Published)
		if err != nil {
			log.Println(err)
			return
		}

		timeSince := time.Since(timeParsed)

		if timeSince < 48*time.Hour {
			log.Println(item.Title, item.Link, timeParsed)
		}
	}

	log.Println("parseFeed done")
}
