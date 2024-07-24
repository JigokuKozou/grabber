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
	"sync"
	"time"

	"github.com/beevik/guid"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Произошла паника: %v", r)
		}
	}()

	start := time.Now()

	sourcePath, destinationPath, err := parseFlags()
	if err != nil {
		log.Fatalln(err)
	}

	if err := createDirectory(destinationPath); err != nil {
		log.Fatalln(err)
	}

	urls, err := readUrlsFromFile(sourcePath)
	if err != nil {
		log.Fatalln(err)
	}

	saveResponses(destinationPath, urls)

	fmt.Printf("Время выполнения: %.2f секунд\n", time.Since(start).Seconds())
}

func saveResponses(destinationPath string, urls []*url.URL) {
	var wg sync.WaitGroup
	wg.Add(len(urls))

	for _, urlSite := range urls {
		go func(destinationPath string, urlSite *url.URL) {
			defer wg.Done()

			err := saveResponse(destinationPath, urlSite)
			if err != nil {
				log.Println(err)
			}
		}(destinationPath, urlSite)
	}

	wg.Wait()
}

// saveResponse - сохранение тела ответа от сервера по указанному URL
func saveResponse(destinationPath string, url *url.URL) error {
	response, err := http.Get(url.String())
	if err != nil {
		return fmt.Errorf("сервер не отвечает [url=%s]: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("не удалось выполнить запрос [status_code=%s, url=%s]", response.Status, url)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения тела ответа [url=%s]: %w", url, err)
	}

	if err := saveResponseData(destinationPath, url, responseBody); err != nil {
		return fmt.Errorf("ошибка сохранения тела ответа [url=%s, destinationPath=%s]: %w", url, destinationPath, err)
	}

	return nil
}

// saveResponseData - сохранение данных тела ответа в файл
func saveResponseData(destinationPath string, url *url.URL, body []byte) error {
	filePath := generateFilePath(destinationPath, url.Host)

	err := os.WriteFile(filePath, body, os.ModePerm)
	if err != nil {
		return fmt.Errorf("ошибка записи тела ответа в файл [url=%s, filePath=%s]: %w", url, filePath, err)
	}
	fmt.Printf("Ответ с сервера \"%s\" успешно записан в файл %s\n", url, filePath)

	return nil
}

// generateFilePath - формирует именя файла с учетом текущего времени и количества запросов к хосту
func generateFilePath(destinationPath string, urlHost string) string {
	fileName := fmt.Sprintf("%s_%s", urlHost, guid.New())
	filePath := filepath.Join(destinationPath, fileName)

	return filePath
}

// readUrlsFromFile - чтение слайса указателей на URL из файла
func readUrlsFromFile(sourcePath string) ([]*url.URL, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла [sourcePath=%s]: %w", sourcePath, err)
	}
	defer file.Close()

	var urls []*url.URL
	var urlCustom *url.URL
	scanner := bufio.NewScanner(file)
	// Чтение URL построчно
	for scanner.Scan() {
		urlCustom, err = url.Parse(scanner.Text())
		if err != nil {
			fmt.Printf("Не корректный URL:'%s'\n", scanner.Text())
			continue
		}
		urls = append(urls, urlCustom)
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// createDirectory - создание директории для сохранения ответов
func createDirectory(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("ошибка создания папки сохранения ответов [path=%s]: %w", path, err)
	}
	return nil
}

// parseFlags - разбор флагов командной строки для получения путей к файлу с URL и папке для сохранения ответов
func parseFlags() (string, string, error) {
	var sourcePath string
	var destinationPath string
	flag.StringVar(&sourcePath, "src", "", "Путь до файла с URL адресами")
	flag.StringVar(&destinationPath, "dst", "", "Путь до папки сохранения ответов")

	flag.Parse()

	if sourcePath == "" || destinationPath == "" {
		flag.Usage()
		return "", "", fmt.Errorf("не указан один из флагов [sourcePath=%s, destinationPath=%s]",
			sourcePath, destinationPath)
	}

	return sourcePath, destinationPath, nil
}
