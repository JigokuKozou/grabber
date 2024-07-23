package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
)

const (
	defaultSourcePath      = "urls.txt"
	defaultDestinationPath = "responses"
)

func main() {
	sourcePath, destinationPath := parseFlags()

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

func readUrlsFromFile(sourcePath string) (urls []*url.URL, err error) {
	fmt.Printf("Считывание из  %s\n", sourcePath)
	testUrl, _ := url.Parse("https://pkg.go.dev/net/url")
	urls = []*url.URL{testUrl}
	return
}

func parseFlags() (sourcePath string, destinationPath string) {
	flag.StringVar(&sourcePath, "src", defaultSourcePath, "Path to file with urls")
	flag.StringVar(&destinationPath, "dst", defaultDestinationPath, "Path to received responses")

	flag.Parse()
	return
}
