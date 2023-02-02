package main

import (
	//"commoncrawl_scraper/parser"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// import "commoncrawl_scraper/parser"

func main() {
	//parser.ParseData("data/indexes/cdx-00000.gz")
	// output := aws_ls("s3://commoncrawl/cc-index/collections/", "2")

	collections := get_collections()
	for _, line := range collections {
		files := get_files(line)
		for _, file := range files {
			file_full_path := fmt.Sprintf("s3://commoncrawl/cc-index/collections/%sindexes/%s", line, file)
			download_file(file_full_path)
			gzip_decompress(filepath.Join("data/indexes/", file))
			break
		}
		break
	}
}

func gzip_decompress(file string) {
	// Create the command
	log.Println("Decompressing file:", file)
	cmd := exec.Command("gzip", "-d", file)

	// Start the command
	cmd.Start()
	cmd.Wait()
}

func download_file(file string) {
	// Create the command
	log.Println("Downloading file:", file)
	cmd := exec.Command("aws", "s3", "cp", file, "data/indexes/")

	// Start the command
	cmd.Start()
	cmd.Wait()
}

func get_files(collection string) []string {
	output := aws_ls(fmt.Sprintf("s3://commoncrawl/cc-index/collections/%sindexes/", collection), "4")
	lines := strings.Split(output, "\n")
	//Remove the last lines, which are empty/unwanted
	if len(lines) > 0 {
		lines = lines[:len(lines)-2]
	}
	return lines
}

func get_collections() []string {
	output := aws_ls("s3://commoncrawl/cc-index/collections/", "2")
	lines := strings.Split(output, "\n")
	// Remove the last lines, which are empty/unwanted
	if len(lines) > 0 {
		lines = lines[:len(lines)-2]
	}
	// Reverse the order of the lines so that newer scrapes are loaded first
	sort.Sort(sort.Reverse(sort.StringSlice(lines)))
	return lines
}

func aws_ls(path string, awk_int string) string {
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
