package parser

import (
	"bufio"
	"commoncrawl_scraper/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type record struct {
	URL          string `json:"url"`
	Mime         string `json:"mime"`
	Status       string `json:"status"`
	Digest       string `json:"digest"`
	Length       string `json:"length"`
	Offset       string `json:"offset"`
	Filename     string `json:"filename"`
	Charset      string `json:"charset"`
	Languages    string `json:"languages"`
	MimeDetected string `json:"mime-detected"`
}

var mimeMap = map[string]string{
	"image/jpeg":      ".jpeg",
	"image/jpg":       ".jpg",
	"image/png":       ".png",
	"application/pdf": ".pdf",
	"video/mp4":       ".mp4",
	"text/csv":        ".csv",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   ".docx",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": "ppt",
}

var mimeLimits = map[string]int64{
	"image/jpeg":      10,
	"image/jpg":       0, // It seems nothing is being detected as jpg, use jpeg instead
	"image/png":       10,
	"application/pdf": 1,
	"video/mp4":       0,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   10,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         1,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": 1,
}

func ParseData(filePath string, mimeTypes []string, maxConcurrent int, c *utils.Container) []string {
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	file, _ := os.Open(filePath)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.Index(line, "{"); i != -1 {
			line = line[i:]
			var r record
			if err := json.Unmarshal([]byte(line), &r); err != nil {
				//fmt.Println("Error parsing JSON:", err)
				continue
			}
			// Validate the URL based on the rules in utils.ValidateURL
			for _, mime := range mimeTypes {
				// check if mime exceeds limit
				mimeTypes = checkMimeTypesLimits(mimeTypes, mime, &wg, c)
				if r.MimeDetected == mime && r.Status == "200" {
					// check if r.Language exists and is not english
					if r.Languages != "" && r.Languages != "eng" {
						log.Println("Mime type", mime, "is not english. Skipping.", r.Languages)
						break
					}
					// start a goroutine to download file. Semaphore is used to limit the number of active goroutines
					wg.Add(1)
					sem <- struct{}{}
					go func(url string, mime string, sem chan struct{}) {
						defer wg.Done()
						downloadFile(url, mime, sem, c)
						<-sem
					}(r.URL, r.MimeDetected, sem)
					break
				}
			}
		}
	}
	return mimeTypes
}

func checkMimeTypesLimits(mimeTypes []string, mime string, wg *sync.WaitGroup, c *utils.Container) []string {
	// if mime exceeds limit remove it from mimeTypes
	c.Mu.RLock()
	if c.DataUsage[mime] > float64(mimeLimits[mime]) {
		for i, m := range mimeTypes {
			if m == mime {
				mimeTypes = append(mimeTypes[:i], mimeTypes[i+1:]...)
				log.Println("Mime type", mime, "exceeded limit. Removing from mimeTypes.")
				break
			}
		}
	}
	c.Mu.RUnlock()
	// if mimeTypes is empty, exit program
	if len(mimeTypes) == 0 {
		log.Println("All mime types exceeded limit. Exiting program.")
		wg.Wait()
		os.Exit(0)
	}
	return mimeTypes
}

func downloadFile(url string, mime string, sem chan struct{}, c *utils.Container) {
	err := utils.ValidateURL(url)
	if err != nil {
		//log.Println(err)
		return
	}
	log.Println("Downloading file:", url)

	//HTTP GET the file
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	defer response.Body.Close()

	// Defer the file name from the URL and use mime type to determine file extension and path
	urlSegments := strings.Split(url, "/")
	fileName := urlSegments[len(urlSegments)-1]
	// Give the file the correct extension
	originalFileExt := filepath.Ext(fileName)
	if originalFileExt == "" {
		fileExt := mimeMap[mime]
		fileName = fileName + fileExt
	} else {
		// strip the file extension and replace it
		strippedFileName := strings.TrimSuffix(fileName, originalFileExt)
		fileExt := mimeMap[mime]
		fileName = strippedFileName + fileExt
	}
	filePath := filepath.Join("data", mime, fileName)
	folderPath := filepath.Join("data", mime)

	// Create the folder if it does not exist
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(folderPath, 0755); err != nil {
			fmt.Println("Error creating folder:", err)
			return
		}
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return
	}

	//write contents to file
	err = ioutil.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Println("Error writing file:", filePath, err)
		return
	}

	// Update data usage
	c.Mu.Lock()
	c.DataUsage[mime] += float64(response.ContentLength) / (1024 * 1024)
	c.Mu.Unlock()
	log.Println("Downloaded file:", filePath)
}
