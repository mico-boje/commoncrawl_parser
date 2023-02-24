package main

import (
	//"commoncrawl_scraper/parser"
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

	// Start the command
	cmd.Start()
	cmd.Wait()
}

func getFiles(collection string) []string {
	output := awsLs(fmt.Sprintf("s3://commoncrawl/cc-index/collections/%sindexes/", collection), "4")
	lines := strings.Split(output, "\n")
	//Remove the last lines, which are empty/unwanted
	if len(lines) > 0 {
		lines = lines[:len(lines)-2]
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
