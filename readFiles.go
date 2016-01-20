package gnul

import (
	"io/ioutil"
	"log"
	"os"
)

var MAX_FILE_SIZE = 10 * 1024 * 1024 // Default 10MB

type Task func()[]ScanResult

func StartReaders(readOrder, outOrder chan *FileInfo, readersCount int) {
	Readers.Add(readersCount)
	for i := 0; i < readersCount; i++ {
		go ReadFiles(readOrder, outOrder)
	}
}

func ReadFiles(in <-chan *FileInfo, out chan<- *FileInfo) {
	// Never die by error
	defer func() {
		err := recover()
		if err != nil {
			log.Println("ReadFiles error (panic): ", err)
			ReadFiles(in, out)
		} else {
			Readers.Done()
		}
	}()

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

func StartScanners(rules []Rule, scanOrder <- chan *FileInfo, resultOrder chan <-ScanResult, scannerCount int){
	tasks := make(Task, scannerCount)
	go Dispatcher(rules, scanOrder, tasks)

	Scanners.Add(scannerCount)
	for i := 0; i < scannerCount; i++{
		go Scanner(tasks, resultOrder)
	}
}

func Dispatcher(rules []Rule, scanOrder <- chan * FileInfo, taskOrder chan <- Task){
	// Never die by error
	defer func(){
		err := recover()
		if err != nil {
			log.Println("Error in Dispatcher: ", err)
			Dispatcher(rules, scanOrder, taskOrder)
		} else {
			close(taskOrder)
		}
	}()

	for file := range scanOrder {
		for _, rule := range rules {
			taskOrder <- func()[]ScanResult{
				return rule(file)
			}
		}
	}
}

func Scanner(taskOrder <- chan Task, resultOrder chan <- ScanResult){
	// Never die by error
	defer func(){
		err := recover()
		if err != nil {
			log.Println("Scanner error (panic): ", err)
			Scanner(taskOrder,resultOrder)
		} else {
			Scanners.Done()
		}
	}()

	for task := range taskOrder {
		scanResults := task()
		for _, res := range scanResults {
			resultOrder <- res
		}
	}
}