package main

import (
	// import standard libraries
	"database/sql"
	"fmt"
	"log"
	"strconv"

	// import third party libraries
	"github.com/PuerkitoBio/goquery"
	_ "github.com/lib/pq"
)

const (
	DB_USER = "maxeinstein"
	DB_NAME = "testdb"
)

func loadContent(page string, done chan<- string) {
	// load html
	doc, err := goquery.NewDocument(page)
	checkErr(err)

	// find relevant bits
	selection := doc.Find("a").Filter(".storylink")
	// collect content from selection
	if selection.Length() > 0 {
		selection.Each(scrapeContent)
	} else {
		done <- "Nothing more to scrape"
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

func main() {
	done := make(chan string)
	base := "http://news.ycombinator.com/news?p="

	for page := 1; page < 100; page++ {
		url := base + strconv.Itoa(page)
		go loadContent(url, done)
	}

	msg := <-done
	fmt.Println(msg)
}
