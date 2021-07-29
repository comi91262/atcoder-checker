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
	"time"

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

func downloadTasks(contestId string) map[string]string {
	url := hostUrl + "/" + path.Join("contests", contestId, "tasks")

	doc, err := downloadDocuments(url)
	if err != nil {
		log.Fatal("fail to download tasks")
		return map[string]string{}
	}

	paths := map[string]string{}
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			td := row.Find("td").First()

			url, _ := td.Find("a").Attr("href")
			if url != "" {
				paths[strings.ToLower(td.Text())] = url
			}
		})
	})

	return paths
}

// taskPath: /contests/abc059/tasks/abc059_b
// ただし, https://atcoder.jp/contests/abc059/tasks/arc072_a
// のように, ARCの問題の一部がABCと共用だった時期があるため、
// ディレクトリは指定した contestId で作る
func downloadSample(contestId, taskId, taskPath string) error {
	doc, err := downloadDocuments(hostUrl + taskPath)
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

	for i := range input {
		fileName := strconv.Itoa(i) + ".txt"
		directoryPath := filepath.Join("sample", contestId, taskId, "in")

		if err := saveFile(input[i], fileName, directoryPath); err != nil {
			log.Fatalf("[error] fail to save %v/%v", directoryPath, fileName)
			return err
		}
	}

	for i := range output {
		fileName := strconv.Itoa(i) + ".txt"
		directoryPath := filepath.Join("sample", contestId, taskId, "out")

		if err := saveFile(output[i], fileName, directoryPath); err != nil {
			log.Fatalf("[error] fail to save %v/%v", directoryPath, fileName)
			return err
		}
	}

	return nil
}

func saveFile(code, fileName, path string) error {
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

func loadFilePath(path string) ([]string, error) {
	paths := []string{}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}

		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})

	return paths, err
}

func downloadSamples(contestId string) {
	taskPaths := downloadTasks(contestId)

	wg := &sync.WaitGroup{}

	for k, v := range taskPaths {
		wg.Add(1)

		taskId, taskPath := k, v
		go func() {
			if err := downloadSample(contestId, taskId, taskPath); err != nil {
				fmt.Printf("[failed] %v\n", taskPath)
			} else {
				fmt.Printf("[success] %v\n", taskPath)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func execute(path string) ([]byte, time.Duration, int64, error) {
	cmd := exec.Command("./main")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
		return []byte{}, 0, 0, err
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return []byte{}, 0, 0, err
	}
	_, err = io.WriteString(stdin, string(bytes))
	if err != nil {
		log.Fatal(err)
		return []byte{}, 0, 0, err
	}
	stdin.Close()

	startTime := time.Now()
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		return []byte{}, 0, 0, err
	}
	elapsedTime := time.Now().Sub(startTime)

	return out, elapsedTime, cmd.ProcessState.SysUsage().(*syscall.Rusage).Maxrss, nil
}

func checkSample(contestId, taskId string) {
	inputs, err1 := loadFilePath(filepath.Join("sample", contestId, taskId, "in"))
	if err1 != nil {
		log.Fatal("[error] failed to load filepath")
		return
	}

	outputs, err2 := loadFilePath(filepath.Join("sample", contestId, taskId, "out"))
	if err2 != nil {
		log.Fatal("[error] failed to load filepath")
		return
	}

	var slowest time.Duration
	var maxMemory int64

	for i := range inputs {
		fmt.Printf("%v\n", inputs[i])

		result, time, memory, err := execute(inputs[i])
		if slowest < time {
			slowest = time
		}
		if maxMemory < memory {
			maxMemory = memory
		}

		expected, err := ioutil.ReadFile(outputs[i])
		if err != nil {
			log.Fatal(err)
			return
		}

		if string(expected) == string(result) {
			fmt.Printf("time: %v \n", time)
			fmt.Println("AC")
		} else {
			fmt.Printf("time: %v \n", time)
			fmt.Println("WA")
			fmt.Println("output:")
			fmt.Printf("%s", result)
			fmt.Println("expected:")
			fmt.Printf("%s", expected)
		}
	}

	fmt.Printf("slowest: %v \n", slowest)
	fmt.Printf("max memory: %v \n", maxMemory)
}

func main() {
	app := cli.App{Name: "atcoder-checker", Usage: "Atcoder utility tools",
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

					checkSample(contestId, taskId)
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
