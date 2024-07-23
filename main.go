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

// saveResponse - сохранение тела ответа от сервера по указанному URL
func saveResponse(destinationPath string, url *url.URL, hostVisitCount map[string]int) error {
	response, err := http.Get(url.String())
	if err != nil {
		return fmt.Errorf("сервер не отвечает URL: \"%s\" Ошибка: %s", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("статус код ответа: \"%s\" URL: \"%s\"", response.Status, url)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения тела ответа URL:\"%s\" Ошибка: %s", url, err)
	}

	if err := saveResponseData(destinationPath, url, hostVisitCount, data); err != nil {
		return err
	}

	return nil
}

// saveResponseData - сохранение данных тела ответа в файл
func saveResponseData(destinationPath string, url *url.URL, hostVisitCount map[string]int, data []byte) error {
	// Формирование имени файла с учетом текущего времени и количества запросов к хосту
	now := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s_%d_%s", url.Host, hostVisitCount[url.Host]+1, now)
	filePath := filepath.Join(destinationPath, fileName)
	err := os.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("ошибка записи данных тела ответа \"%s\" в файл \"%s\" Ошибка: %s", url, filePath, err)
	}

	// Увеличение счетчика запросов к хосту
	hostVisitCount[url.Host]++
	return nil
}

// readUrlsFromFile - чтение слайса указателей на URL из файла
func readUrlsFromFile(sourcePath string) ([]*url.URL, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла \"%s\" Ошибка: %s", sourcePath, err)
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
		return fmt.Errorf("ошибка создания папки сохранения ответов по пути: \"%s\" Ошибка: \"%s\"", path, err)
	}
	return nil
}

// parseFlags - разбор флагов командной строки для получения путей к файлу с URL и папке для сохранения ответов

func parseFlags() (string, string) {
	var sourcePath string
	var destinationPath string
	flag.StringVar(&sourcePath, "src", defaultSourcePath, "Путь до файла с URL адресами")
	flag.StringVar(&destinationPath, "dst", defaultDestinationPath, "Путь до папки сохранения ответов")

	flag.Parse()
	return sourcePath, destinationPath
}
