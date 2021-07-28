package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli/v2"
)

var reader = bufio.NewReader(os.Stdin)
var writer = bufio.NewWriter(os.Stdout)

const hostUrl = "https://atcoder.jp"

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

	url := hostUrl + "/" + path.Join("contests", contestId, "tasks")

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

func downloadSample(taskUrl string) {

	url := hostUrl + taskUrl
	fmt.Printf("%v\n", url)

	doc, err := downloadDocuments(url)
	if err != nil {
		log.Fatal("fail to download sample")
		return
	}

	input := []string{}
	output := []string{}
	doc.Find("section").Each(func(i int, s *goquery.Selection) {
		if strings.HasPrefix(s.Find("h3").Text(), "入力例") {
			fmt.Printf("%v\n", s.Find("h3").Text())

			text := ""
			s.Find("pre").Each(func(i int, s *goquery.Selection) {
				fmt.Printf("%v\n", s.Text())
				text += s.Text() + "\n"
			})
			input = append(input, text)
		}
		if strings.HasPrefix(s.Find("h3").Text(), "出力例") {
			fmt.Printf("%v\n", s.Find("h3").Text())

			text := ""
			s.Find("pre").Each(func(i int, s *goquery.Selection) {
				fmt.Printf("%v\n", s.Text())
				text += s.Text() // + "\n"
			})
			output = append(output, text)
		}
	})

	fmt.Fprintf(writer, "%v\n", input)
	fmt.Fprintf(writer, "%v\n", output)
	fmt.Fprintf(writer, "%v\n", taskUrl)
	writer.Flush()
	paths := strings.Split(taskUrl, "/")
	fmt.Fprintf(writer, "%v\n", paths)
	writer.Flush()
	ids := strings.Split(paths[len(paths)-1], "_")
	contestId := ids[0]
	taskId := ids[1]

	for i := range input {
		archiveFile(input[i], strconv.Itoa(i)+"txt", filepath.Join("sample", contestId, taskId, "in"))
	}

	for i := range output {
		archiveFile(output[i], strconv.Itoa(i)+"txt", filepath.Join("sample", contestId, taskId, "out"))
	}
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
					contestId := c.Args().Get(0)
					// fmt.Printf("%v\n", contestId)
					// fmt.Printf("%v\n", downloadTasks(contestId))
					taskUrls := downloadTasks(contestId)

					fmt.Printf("%v\n", len(taskUrls))
					for i := range taskUrls {
						fmt.Printf("%v\n", i)
						fmt.Printf("%v\n", taskUrls[i])
						if taskUrls[i] == "" {
							continue
						}
						downloadSample(taskUrls[i])
					}
					// downloadSample("abc010", "abc010_1")
					// downloadSample("abc057", "abc057_b")
					// downloadSample("abc100", "abc100_a")
					// downloadSample("abc200", "abc200_a")
					// execute()
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
