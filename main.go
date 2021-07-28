package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

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

	inputText := ""
	outputText := ""
	doc.Find("section").Each(func(i int, s *goquery.Selection) {
		if strings.HasPrefix(s.Find("h3").Text(), "入力例") {
			fmt.Printf("%v\n", s.Find("h3").Text())
			s.Find("pre").Each(func(i int, s *goquery.Selection) {
				fmt.Printf("%v\n", s.Text())
				inputText += s.Text() + "\n"
			})
		}
		if strings.HasPrefix(s.Find("h3").Text(), "出力例") {
			fmt.Printf("%v\n", s.Find("h3").Text())
			s.Find("pre").Each(func(i int, s *goquery.Selection) {
				fmt.Printf("%v\n", s.Text())
				outputText += s.Text() + "\n"
			})
		}
	})

	archiveFile(inputText, "input.txt", "sample")
	archiveFile(outputText, "output.txt", "sample")
	//fmt.Println(string(byteArray)) // htmlをstringで取得
}

// func loadSample() {
// 	filepath.Walk("sample", func(path string, info os.FileInfo, err error) error {
// 		if !info.IsDir() {
// 			fmt.Printf("%v\n", path)
// 			fp, err := os.Open(path)
// 			if err != nil {
// 				log.Println(err)
// 				return err
// 			}
// 			defer fp.Close()
// 			scanner := bufio.NewScanner(fp)
// 			for scanner.Scan() {
// 				archivedKeys[scanner.Text()] = struct{}{}
// 			}
// 		}
// 		return nil
// 	})
// }

func makeDirectory(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func archiveFile(code, fileName, path string) error {
	makeDirectory(path)
	filePath := filepath.Join(path, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(code)
	return nil
}

// func compile() {
// 	// TODO not found hoge.go
// 	out, err := exec.Command("go", "build", "-o", "main", "hoge.go").Output()
// 	if err != nil {
// 		log.Fatal(err)
// 		return
// 	}
//
// 	out, err := exec.Command("go", "build", "-o", "main", "hoge.go").Output()
// 	fmt.Printf("結果: %s", out)
// }

func execute() {
	cmd := exec.Command("./hoge")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
		return
	}

	bytes, err := ioutil.ReadFile("sample/input.txt")
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.WriteString(stdin, string(bytes))
	if err != nil {
		log.Fatal(err)
	}
	stdin.Close()

	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("%s", out)
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
					// fmt.Print(downloadTasks("abc001"))
					// downloadSample("abc001", "abc001_1")
					// downloadSample("abc010", "abc010_1")
					// downloadSample("abc057", "abc057_b")
					// downloadSample("abc100", "abc100_a")
					// downloadSample("abc200", "abc200_a")
					execute()
					return nil
				},
			},
			{
				Name:    "compile",
				Aliases: []string{"c"},
				Usage:   "compile program sample",
				Action: func(c *cli.Context) error {
					//	compile()
					execute()
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
