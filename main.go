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
)

const (
	defaultSourcePath      = "urls.txt"
	defaultDestinationPath = "responses"
)

func main() {
	start := time.Now()

	sourcePath, destinationPath := parseFlags()

	if err := createDirectory(destinationPath); err != nil {
		log.Fatalln(err)
	}

	urls, err := readUrlsFromFile(sourcePath)
	if err != nil {
		log.Fatalln(err)
	}

	saveResponses(destinationPath, urls)

	fmt.Printf("Время выполнения: %s\n", time.Since(start))
}

func saveResponses(destinationPath string, urls []*url.URL) {
	hostVisits := make(map[string]int)
	var hostVisitsMx sync.RWMutex

	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, urlSite := range urls {
		go func(destinationPath string, urlSite *url.URL, hostVisits map[string]int, hostVisitsMx *sync.RWMutex) {
			defer wg.Done()

			err := saveResponse(destinationPath, urlSite, hostVisits, hostVisitsMx)
			if err != nil {
				log.Println(err)
			}
		}(destinationPath, urlSite, hostVisits, &hostVisitsMx)
	}
	wg.Wait()
}

// saveResponse - сохранение тела ответа от сервера по указанному URL
func saveResponse(destinationPath string, url *url.URL, hostVisits map[string]int, hostVisitsMx *sync.RWMutex) error {
	response, err := http.Get(url.String())
	if err != nil {
		return fmt.Errorf("сервер не отвечает [url=%s]: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("не удалось выполнить запрос [status_code=%s, url=%s]", response.Status, url)
	}

	responseBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения тела ответа [url=%s]: %w", url, err)
	}

	if err := saveResponseData(destinationPath, url, responseBodyBytes, hostVisits, hostVisitsMx); err != nil {
		return err
	}

	return nil
}

// saveResponseData - сохранение данных тела ответа в файл
func saveResponseData(destinationPath string, url *url.URL, bytes []byte, hostVisits map[string]int, hostVisitsMx *sync.RWMutex) error {
	hostVisitsMx.RLock()
	filePath := generateFilePath(destinationPath, url.Host, hostVisits[url.Host]+1)
	hostVisitsMx.RUnlock()

	err := os.WriteFile(filePath, bytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("ошибка записи данных тела ответа в файл [url=%s, filePath=%s]: %w", url, filePath, err)
	}

	// Увеличение счетчика запросов к хосту
	hostVisitsMx.Lock()
	hostVisits[url.Host]++
	hostVisitsMx.Unlock()

	return nil
}

// generateFilePath - формирует именя файла с учетом текущего времени и количества запросов к хосту
func generateFilePath(destinationPath string, urlHost string, countVisits int) string {
	now := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s_%d_%s", urlHost, countVisits, now)
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
func parseFlags() (string, string) {
	var sourcePath string
	var destinationPath string
	flag.StringVar(&sourcePath, "src", defaultSourcePath, "Путь до файла с URL адресами")
	flag.StringVar(&destinationPath, "dst", defaultDestinationPath, "Путь до папки сохранения ответов")

	flag.Parse()

	if sourcePath == defaultSourcePath && destinationPath == defaultDestinationPath {
		flag.Usage()
	}

	return sourcePath, destinationPath
}
