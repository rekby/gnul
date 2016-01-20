package gnul

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
)

var MAX_FILE_SIZE int64 = 10 * 1024 * 1024 // Default 10MB

type Task func() []ScanResult

func StartReaders(readOrder, outOrder chan *FileInfo, readersCount int) {
	Readers.Add(readersCount)
	for i := 0; i < readersCount; i++ {
		go func() {
			log.Println("Start reader")
			ReadFiles(readOrder, outOrder)
			Readers.Done()
			log.Println("Close reader")
		}()
	}
}

func ReadFiles(in <-chan *FileInfo, out chan<- *FileInfo) {
	for file := range in {
		stat, err := os.Stat(file.Path)
		if err != nil {
			log.Printf("Can't stat file: '%v', err: '%v'\n", file.Path, err)
			continue
		}
		if stat.Size() > MAX_FILE_SIZE {
			log.Println("Skip by size: ", file.Path)
			continue
		}
		file.content, err = ioutil.ReadFile(file.Path)
		if err != nil {
			log.Printf("Can't read file: '%v', err: '%v'", file.Path, err)
			continue
		}
		out <- file
	}
}

// It close readOrder
func ReadFilesFromArgs(readOrder chan<- *FileInfo, files ...string) {
	defer close(readOrder)
	for _, filePath := range files {
		readOrder <- &FileInfo{Path: filePath}
	}
}

// It block until EOF in stdin or any error, then close readOrder
func ReadFilesFromStdIn(readOrder chan<- *FileInfo) {
	defer close(readOrder)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		readOrder <- &FileInfo{Path: line}
	}
	if scanner.Err() != nil {
		log.Println("Error while read filenames from stdin:", scanner.Err())
	}
}
