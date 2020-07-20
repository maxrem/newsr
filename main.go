package main

import (
	"database/sql"
	"fmt"
	"hash/fnv"
	"log"
	"net/url"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
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

	googleNewsUrl := viper.GetString("google.news-url")
	dateTimeLayout := viper.GetString("google.date-time-layout")
	keyWord := viper.GetString("google.keyword")
	workerCount := viper.GetInt("worker-count")

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
				parseFeed(u, resultChannel, dateTimeLayout)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	insertCount := 0
	for res := range resultChannel {
		result, err := dbConn.Exec(
			"INSERT IGNORE INTO `article` (`title`, `url`, `hash`, `description`, `content`, `published`) VALUES (?, ?, ?, ?, ?, ?)",
			res.Title,
			res.Url,
			res.Hash,
			res.Description,
			res.Content,
			res.Published,
		)
		if err != nil {
			log.Println(err)
		} else {
			id, err := result.LastInsertId()
			if err == nil && id > 0 {
				insertCount++
			}
		}
	}

	log.Println("inserted", insertCount, "into database")
}

func parseFeed(url string, resultChannel chan Article, dateTimeLayout string) {
	log.Println("parsing", url)
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(url)

	if err != nil {
		log.Println(err)
		return
	}

	articleCount := 0
	for i := 0; i < feed.Len(); i++ {
		item := feed.Items[i]

		publishedTime, err := time.Parse(dateTimeLayout, item.Published)
		if err != nil {
			log.Println(err)
			return
		}

		timeSince := time.Since(publishedTime)

		if timeSince < 48*time.Hour {
			articleCount++
			resultChannel <- Article{
				Title:       item.Title,
				Url:         item.Link,
				Hash:        hash(item.Link),
				Description: item.Description,
				Published:   publishedTime,
			}
		}
	}

	log.Println("parsing done. found", articleCount, "new articles")
}

func hash(s string) uint64 {
	h := fnv.New64a()
	_, err := h.Write([]byte(s))
	if err != nil {
		log.Println(err)
	}
	return h.Sum64()
}
