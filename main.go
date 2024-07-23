package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
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

	for _, url := range urls {
		err := saveResponse(destinationPath, url)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func saveResponse(destinationPath string, url *url.URL) error {
	fmt.Printf("Сохранение ответа в %s от %s\n", destinationPath, url)

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
