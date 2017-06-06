package main

import (
	// import standard libraries
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"

	// import third party libraries
	"github.com/PuerkitoBio/goquery"
	_ "github.com/lib/pq"
)

const (
	DB_USER = "maxeinstein"
	DB_NAME = "testdb"
)

func main() {
	var wg sync.WaitGroup
	done := make(chan bool)
	base := "http://news.ycombinator.com/news?p="
	page := 1

Loop:
	for {
		select {
		case endScraper := <-done:
			_ = endScraper
			break Loop
		default:
			// increment page number
			page++
			// print page number
			fmt.Println("page-->", page)
			// construct next page url
			url := base + strconv.Itoa(page)
			// add one to sync group
			wg.Add(1)
			// spawn go rountine to load, scrape and save data
			go loadContent(url, done, &wg)
		}
	}

	wg.Wait()
	fmt.Println("wg.Wait() done")
}

func loadContent(url string, done chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	// load html
	doc, err := goquery.NewDocument(url)
	checkErr(err)

	// find relevant bits
	selection := doc.Find("a").Filter(".storylink")
	// collect content from selection
	if selection.Length() > 0 {
		selection.Each(scrapeContent)
	} else {
		done <- true
	}
}

func scrapeContent(index int, item *goquery.Selection) {
	// grab the title
	title := item.Text()

	// print title
	fmt.Printf("Article #%d: %s\n", index, title)

	dbinfo := fmt.Sprintf("user=%s dbname=%s sslmode=disable", DB_USER, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err.Error())
	}

	// prepare postgres statement
	sStmt := "INSERT INTO test(title) VALUES($1)"

	stmt, err := db.Prepare(sStmt)
	checkErr(err)

	res, err := stmt.Exec(title)
	checkErr(err)
	// ignore res
	_ = res
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
