package main

import (
	//"commoncrawl_scraper/parser"
	"bufio"
	"bytes"
	"commoncrawl_scraper/parser"
	"commoncrawl_scraper/utils"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	c := utils.Container{Mu: sync.RWMutex{}, DataUsage: make(map[string]float64)}
	allowedMimes := []string{"image/png", "image/jpeg", "application/pdf", "video/mp4", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/vnd.openxmlformats-officedocument.presentationml.presentation"}
	var threads int
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <number of threads>")
		return
	}
	threads, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid number of threads")
		return
	}
	log.Printf("Number of threads: %d\n", threads)

	collections := getCollections()
	for _, line := range collections {
		files := getFiles(line)
		for _, file := range files {
			file_full_path := fmt.Sprintf("s3://commoncrawl/cc-index/collections/%sindexes/%s", line, file)
			downloadFile(file_full_path)
			gzipDecompress(filepath.Join("data/indexes/", file))
			file = filepath.Join("data/indexes/", strings.TrimSuffix(file, filepath.Ext(file)))
			log.Println("Parsing file:", file)
			allowedMimes = parser.ParseData(file, allowedMimes, threads, &c)
			removeFile(file)
		}
	}
}

func removeFile(file string) {
	log.Println("Removing file:", file)
	err := os.Remove(file)
	if err != nil {
		log.Println(err)
	}
}

func gzipDecompress(file string) {
	// Create the command
	log.Println("Decompressing file:", file)
	cmd := exec.Command("gzip", "-d", file)

	// Start the command
	cmd.Start()
	cmd.Wait()
}

func downloadFile(file string) {
	// Create the command
	log.Println("Downloading file:", file)
	cmd := exec.Command("aws", "s3", "cp", file, "data/indexes/")

	// Create a buffer to hold the output
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Start the command
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the command to finish
	err = cmd.Wait()

	// If there was an error and it's a SlowDown error, sleep and retry
	if err != nil {
		if strings.Contains(stderr.String(), "An error occurred (SlowDown) when calling the GetObject operation") {
			for i := 0; i < 4; i++ {
				log.Printf("Retrying download in %d seconds...\n", (i+1)*60)
				time.Sleep(time.Duration((i+1)*60) * time.Second)
				err = cmd.Start()
				if err != nil {
					log.Fatal(err)
				}
				err = cmd.Wait()
				if err == nil {
					log.Println("Download successful.")
					return
				}
				if i == 3 {
					log.Fatal("Max retries exceeded.")
				}
			}
		} else {
			log.Fatal(stderr.String())
		}
	}

	log.Println("Download successful.")
}

func getFiles(collection string) []string {
	output := awsLs(fmt.Sprintf("s3://commoncrawl/cc-index/collections/%sindexes/", collection), "4")
	lines := strings.Split(output, "\n")
	// Remove the last lines, which are empty/unwanted
	if len(lines) > 0 {
		lines = lines[:len(lines)-2]
	}
	// Keep only lines from index 30 to len(lines)-2
	if len(lines) > 30 {
		lines = lines[30 : len(lines)-2]
	} else {
		lines = []string{} // return an empty slice if there are less than 30 lines
	}
	return lines
}

func getCollections() []string {
	output := awsLs("s3://commoncrawl/cc-index/collections/", "2")
	lines := strings.Split(output, "\n")
	// Remove the last lines, which are empty/unwanted
	if len(lines) > 0 {
		lines = lines[:len(lines)-2]
	}
	// Remove lines if they are present in indexes.txt file
	indexes, err := os.Open("indexes.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer indexes.Close()
	scanner := bufio.NewScanner(indexes)
	for scanner.Scan() {
		for i, line := range lines {
			if line == scanner.Text() {
				lines = append(lines[:i], lines[i+1:]...)
			}
		}
	}

	// Reverse the order of the lines so that newer scrapes are loaded first
	sort.Sort(sort.Reverse(sort.StringSlice(lines)))
	return lines
}

func awsLs(path string, awk_int string) string {
	// Create the command
	cmd := exec.Command("aws", "s3", "ls", path)

	// Create a new command to run awk and pass the output from the first command to it
	awk := exec.Command("awk", fmt.Sprintf("{print $%s}", awk_int))

	// Set the stdout of the first command to be the stdin of the second command
	awk.Stdin, _ = cmd.StdoutPipe()

	// Capture the output of the second command
	var output bytes.Buffer
	awk.Stdout = &output

	// Start both commands
	cmd.Start()
	awk.Start()
	cmd.Wait()
	awk.Wait()

	return output.String()
}
