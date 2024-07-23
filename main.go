package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultSourcePath      = "urls.txt"
	defaultDestinationPath = "responses"
)

func main() {
	sourcePath, destinationPath := parseFlags()

	if err := createDirectory(destinationPath); err != nil {
		log.Fatalln(err)
	}

	urls, err := readUrlsFromFile(sourcePath)
	if err != nil {
		log.Fatalln(err)
	}

	urlHostCounter := make(map[string]int)
	for _, url := range urls {
		err := saveResponse(destinationPath, url, urlHostCounter)
		if err != nil {
			log.Println(err)
		}
	}
}

func saveResponse(destinationPath string, url *url.URL, urlHostCounter map[string]int) error {
	response, err := http.Get(url.String())
	if err != nil {
		return fmt.Errorf("server is not responding URL: \"%s\" Error: %s", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server status code: \"%s\" URL: \"%s\"", response.Status, url)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := saveResponseData(destinationPath, url, urlHostCounter, data); err != nil {
		return err
	}

	return nil
}

func saveResponseData(destinationPath string, url *url.URL, urlHostCounter map[string]int, data []byte) error {
	now := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s_%d_%s", url.Host, urlHostCounter[url.Host]+1, now)
	filePath := filepath.Join(destinationPath, fileName)
	err := os.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		return err
	}

	urlHostCounter[url.Host]++
	return nil
}

func readUrlsFromFile(sourcePath string) ([]*url.URL, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []*url.URL
	var urlCustom *url.URL
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urlCustom, err = url.Parse(scanner.Text())
		if err != nil {
			fmt.Printf("Invalid url:'%s'\n", scanner.Text())
			continue
		}
		urls = append(urls, urlCustom)
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func createDirectory(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func parseFlags() (sourcePath string, destinationPath string) {
	flag.StringVar(&sourcePath, "src", defaultSourcePath, "Path to file with urls")
	flag.StringVar(&destinationPath, "dst", defaultDestinationPath, "Path to received responses")

	flag.Parse()
	return
}
