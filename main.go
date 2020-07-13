package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
	"hash/fnv"
	"log"
	"net/url"
	"sync"
	"time"
)

type Article struct {
	Id          int64
	Title       string
	Url         string
	Hash        uint64
	Description string
	Content     string
	Published   time.Time
}

const googleNewsUrl = "https://news.google.com/news?hl=en-US&ceid=US:en&sort=date&gl=US&num=10&output=rss&q="
const dateTimeLayout = "Mon, 02 Jan 2006 15:04:05 MST"
const keyWord = "covid"
const workerCount = 4

func main() {
	var wg sync.WaitGroup

	urlChannel := make(chan string)
	resultChannel := make(chan Article)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	dbConn, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(localhost:13306)/%s",
			viper.GetString("mysql.username"),
			viper.GetString("mysql.password"),
			viper.GetString("mysql.db_name"),
		))
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

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
		_, err := dbConn.Query(
			"INSERT INTO `article` (`title`, `url`, `hash`, `description`, `content`, `published`) VALUES (?, ?, ?, ?, ?, ?)",
			res.Title,
			res.Url,
			res.Hash,
			res.Description,
			res.Content,
			res.Published,
		)
		if err != nil {
			log.Println(err)
		}
	}
}

func parseFeed(url string, resultChannel chan Article) {
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
			resultChannel <- Article{
				Title:       item.Title,
				Url:         item.Link,
				Hash:        hash(item.Link),
				Description: item.Description,
				Published:   publishedTime,
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
