package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli/v2"
)

const contestUrl = "https://atcoder.jp/contests"

func downloadDocuments(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return nil, err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc, nil
}

func downloadTasks(contestId string) []string {

	url := contestUrl + "/" + path.Join(contestId, "tasks")

	doc, err := downloadDocuments(url)
	if err != nil {
		log.Fatal("fail to download html")
		return []string{}
	}

	paths := []string{}
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			url, _ := row.Find("td").Next().Find("a").Attr("href")
			paths = append(paths, url)
		})
	})

	return paths
}

func downloadSample(contestId, problemId string) {

	url := contestUrl + "/" + path.Join(contestId, "tasks", problemId)

	doc, err := downloadDocuments(url)
	if err != nil {
		log.Fatal("fail to download sample")
		return
	}

	tex := ""
	doc.Find("section").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("%v\n", s.Find("h3").Text())
		s.Find("pre").Each(func(i int, s *goquery.Selection) {
			fmt.Printf("%v\n", s.Text())
			tex += s.Text()
		})
	})

	archiveFile(tex, "a.txt", "sample")

	//fmt.Println(string(byteArray)) // htmlをstringで取得
}

func archiveFile(code, fileName, path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	filePath := filepath.Join(path, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(code)
	return nil
}

//
//  // Find the review items
//  doc.Find(".left-content article .post-title").Each(func(i int, s *goquery.Selection) {
//		// For each item found, get the title
//		title := s.Find("a").Text()
//		fmt.Printf("Review %d: %s\n", i, title)
//	})

func main() {
	app := cli.App{Name: "oj-go", Usage: "Atcoder utility tools",
		Commands: []*cli.Command{
			{
				Name:    "download",
				Aliases: []string{"d"},
				Usage:   "download sample",
				Action: func(c *cli.Context) error {
					fmt.Print(downloadTasks("abc001"))
					fmt.Print(downloadTasks("abc010"))
					fmt.Print(downloadTasks("abc100"))
					fmt.Print(downloadTasks("abc190"))
					downloadSample("abc046", "abc046_a")
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
