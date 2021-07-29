package main

import (
	"errors"
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
	"sync"
	"syscall"

	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli/v2"
)

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
		log.Fatal("fail to download tasks")
		return []string{}
	}

	paths := []string{}
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			url, _ := row.Find("td").Next().Find("a").Attr("href")
			if url != "" {
				paths = append(paths, url)
			}
		})
	})

	return paths
}

func downloadSample(taskUrl string) error {
	doc, err := downloadDocuments(hostUrl + taskUrl)
	if err != nil {
		log.Fatal("[error] fail to download sample")
		return err
	}

	input := []string{}
	output := []string{}
	doc.Find("section").Each(func(i int, s *goquery.Selection) {
		if strings.HasPrefix(s.Find("h3").Text(), "入力例") {
			text := ""
			s.Find("pre").Each(func(i int, s *goquery.Selection) {
				text += s.Text()
			})
			input = append(input, text)
		}
		if strings.HasPrefix(s.Find("h3").Text(), "出力例") {
			text := ""
			s.Find("pre").Each(func(i int, s *goquery.Selection) {
				text += s.Text()
			})
			output = append(output, text)
		}
	})

	paths := strings.Split(taskUrl, "/")
	ids := strings.Split(paths[len(paths)-1], "_")
	contestId, taskId := ids[0], ids[1]

	for i := range input {
		fileName := strconv.Itoa(i) + ".txt"
		directoryPath := filepath.Join("sample", contestId, taskId, "in")

		if err := archiveFile(input[i], fileName, directoryPath); err != nil {
			log.Fatalf("[error] fail to save %v/%v", directoryPath, fileName)
			return err
		}
	}

	for i := range output {
		fileName := strconv.Itoa(i) + ".txt"
		directoryPath := filepath.Join("sample", contestId, taskId, "out")

		if err := archiveFile(output[i], fileName, directoryPath); err != nil {
			log.Fatalf("[error] fail to save %v/%v", directoryPath, fileName)
			return err
		}
	}

	return nil
}

func archiveFile(code, fileName, path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		log.Fatal(err)
		return err
	}

	file, err := os.Create(filepath.Join(path, fileName))
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(code); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func execute(contestId, taskId string) {
	inputPath := filepath.Join("sample", contestId, taskId, "in")

	inputs := []string{}
	filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return nil
		}

		if !info.IsDir() {
			inputs = append(inputs, path)
		}
		return nil
	})

	outputs := []string{}
	outputPath := filepath.Join("sample", contestId, taskId, "out")
	filepath.Walk(outputPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			outputs = append(outputs, path)
		}
		return nil
	})

	for i := range inputs {
		cmd := exec.Command("./hoge")

		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
			return
		}

		bytes, err := ioutil.ReadFile(inputs[i])
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
		//fmt.Printf("%v\n", cmd.ProcessState.SysUsage().(*syscall.Rusage).Maxrss)
		fmt.Printf("%v\n", cmd.ProcessState.SysUsage().(*syscall.Rusage).Maxrss)
		fmt.Printf("%v\n", cmd.ProcessState.SystemTime())
		fmt.Printf("%v\n", cmd.ProcessState.UserTime())
		//startTime := time.Now()
		//elapsedTime := time.Now().Sub(startTime)

		bytes, err = ioutil.ReadFile(outputs[i])
		if err != nil {
			panic(err)
		}

		if string(bytes) == string(out) {
			fmt.Println("AC")
		} else {
			fmt.Println("WA")
			fmt.Printf("%s", bytes)
			fmt.Printf("%s", out)
		}
	}

}

func downloadSamples(contestId string) {
	taskUrls := downloadTasks(contestId)

	wg := &sync.WaitGroup{}

	for i := range taskUrls {
		wg.Add(1)
		idx := i
		go func() {
			if err := downloadSample(taskUrls[idx]); err != nil {
				fmt.Printf("[failed] %v\n", taskUrls[idx])
			} else {
				fmt.Printf("[success] %v\n", taskUrls[idx])
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func main() {
	app := cli.App{Name: "oj-go", Usage: "Atcoder utility tools",
		Commands: []*cli.Command{
			{
				Name:    "download",
				Aliases: []string{"d"},
				Usage:   "download sample",
				Action: func(c *cli.Context) error {
					contestId := c.Args().Get(0)
					if contestId == "" {
						return errors.New("[error] contestId is required e.g abc001")
					}
					downloadSamples(contestId)
					return nil
				},
			},
			{
				Name:    "check",
				Aliases: []string{"c"},
				Usage:   "check sample",
				Action: func(c *cli.Context) error {
					contestId := c.Args().Get(0)
					if contestId == "" {
						return errors.New("[error] contestId is required e.g abc001")
					}
					taskId := c.Args().Get(1)
					if taskId == "" {
						return errors.New("[error] taskId is required e.g a or 1")
					}

					execute(contestId, taskId)
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
