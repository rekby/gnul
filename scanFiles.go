package gnul

import (
	"sync"
)

// Block until stop all scanners
func StartScanners(rules []Rule, scanOrder <-chan *FileInfo, resultOrder chan<- ScanResult, scannerCount int) {
	tasks := make(chan Task, scannerCount)
	go Dispatcher(rules, scanOrder, tasks)

	var scanners sync.WaitGroup
	scanners.Add(scannerCount)
	for i := 0; i < scannerCount; i++ {
		go func() {
			Scanner(tasks, resultOrder)
			scanners.Done()
		}()
	}
	scanners.Wait()
}

func Dispatcher(rules []Rule, scanOrder <-chan *FileInfo, taskOrder chan<- Task) {
	for file := range scanOrder {
		for _, rule := range rules {
			taskOrder <- func() []ScanResult {
				return rule(file)
			}
		}
	}
	close(taskOrder)
}

func Scanner(taskOrder <-chan Task, resultOrder chan<- ScanResult) {
	for task := range taskOrder {
		scanResults := task()
		for _, res := range scanResults {
			resultOrder <- res
		}
	}
}
