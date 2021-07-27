package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	url := "https://atcoder.jp/contests/abc046/tasks/abc046_a"

	res, _ := http.Get(url)
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("section").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("%v\n", s.Find("h3").Text())
		s.Find("pre").Each(func(i int, s *goquery.Selection) {
			fmt.Printf("%v\n", s.Text())
		})
	})

	//fmt.Println(string(byteArray)) // htmlをstringで取得
}

//res, err := http.Get("http://metalsucks.net")
//
//
//  if err != nil {
//    log.Fatal(err)
//  }
//  defer res.Body.Close()
//  if res.StatusCode != 200 {
//    log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
//  }
//
//  // Load the HTML document
//  doc, err := goquery.NewDocumentFromReader(res.Body)
//  if err != nil {
//    log.Fatal(err)
//  }
//
//  // Find the review items
//  doc.Find(".left-content article .post-title").Each(func(i int, s *goquery.Selection) {
//		// For each item found, get the title
//		title := s.Find("a").Text()
//		fmt.Printf("Review %d: %s\n", i, title)
//	})
