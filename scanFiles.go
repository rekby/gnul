package gnul

import "log"

func StartScanners(rules []Rule, scanOrder <-chan *FileInfo, resultOrder chan<- ScanResult, scannerCount int) {
	tasks := make(chan Task, scannerCount)
	go Dispatcher(rules, scanOrder, tasks)

	Scanners.Add(scannerCount)
	for i := 0; i < scannerCount; i++ {
		go func() {
			log.Println("Start scanner")
			Scanner(tasks, resultOrder)
			log.Println("Close scanner")
			Scanners.Done()
		}()
	}
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
