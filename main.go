package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	url := "https://atcoder.jp/contests/abc046/tasks/abc046_a"

	resp, _ := http.Get(url)
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(byteArray)) // htmlをstringで取得
}

//res, err := http.Get("http://metalsucks.net")
//
//c
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
