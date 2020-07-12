package main

import (
	"bytes"
	"log"
	"net/http"
)

var googleNewsUrl = "http://news.google.com/news?hl=en-US&sort=date&gl=US&num=10&output=rss&q="

func main() {
	keyWord := "covid"

	downloadHtml(googleNewsUrl + keyWord)

	log.Println()
}

func downloadHtml(url string) (contents string, err error) {
	resp, err := http.Get(url)

	if err != nil {
		return
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return
	}
	contents = buf.String()

	return
}
