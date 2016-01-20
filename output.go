package gnul

import "fmt"

// It block until resultOrder close
func PrintFileNameAndRule(results <-chan ScanResult) {
	for res := range results {
		fmt.Println(res.File.Path, res.Title)
	}
}