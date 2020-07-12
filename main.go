package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"hash/fnv"
	"log"
	"net/url"
	"sync"
	"time"
)

type ParseResult struct {
	title       string
	url         string
	hash        uint64
	description string
	content     string
	published   time.Time
}

const googleNewsUrl = "https://news.google.com/news?hl=en-US&ceid=US:en&sort=date&gl=US&num=10&output=rss&q="
const dateTimeLayout = "Mon, 02 Jan 2006 15:04:05 MST"
const keyWord = "covid"
const workerCount = 4

// TODO

func main() {
	var wg sync.WaitGroup

	urlChannel := make(chan string)
	resultChannel := make(chan ParseResult)

	subjectList := []string{
		"10x Genomics",
		"89bio",
		"Abeona Therapeutics Inc.",
		"AC Immune SA",
		"ACADIA Pharmaceuticals Inc.",
		"Acasti Pharma, Inc.",
		"Accelerate Diagnostics, Inc.",
		"Acceleron Pharma Inc.",
		"AcelRx Pharmaceuticals, Inc.",
		"Acer Therapeutics Inc.",
		"Achaogen, Inc.",
		"Achieve Life Sciences, Inc.",
		"Aclaris Therapeutics, Inc.",
		"Acorda Therapeutics, Inc.",
		"Adamas Pharmaceuticals, Inc.",
		"Adamis Pharmaceuticals Corporation",
		"Adaptimmune Therapeutics plc",
		"Adaptive Biotechnologies",
		"Adial Pharmaceuticals",
		"ADMA Biologics Inc",
	}

	go func() {
		for _, subject := range subjectList {
			urlChannel <- googleNewsUrl + fmt.Sprintf("%s+%s", url.QueryEscape(keyWord), url.QueryEscape(subject))
		}

		close(urlChannel)
	}()

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range urlChannel {
				parseFeed(u, resultChannel)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	for res := range resultChannel {
		fmt.Println(res)
	}
}

func parseFeed(url string, resultChannel chan ParseResult) {
	log.Println("Parsing", url)
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(url)

	if err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < feed.Len(); i++ {
		item := feed.Items[i]

		publishedTime, err := time.Parse(dateTimeLayout, item.Published)
		if err != nil {
			log.Println(err)
			return
		}

		timeSince := time.Since(publishedTime)

		if timeSince < 48*time.Hour {
			resultChannel <- ParseResult{
				title:       item.Title,
				url:         item.Link,
				hash:        hash(item.Link),
				description: item.Description,
				published:   publishedTime,
			}
		}
	}

	log.Println("Parsing done")
}

func hash(s string) uint64 {
	h := fnv.New64a()
	_, err := h.Write([]byte(s))
	if err != nil {
		log.Println(err)
	}
	return h.Sum64()
}
