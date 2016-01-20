package gnul

import (
	"encoding/xml"
	"fmt"
	"github.com/rekby/pflag"
	"io/ioutil"
	"log"
	"runtime"
	"sync"
)

//go:generate stringer -type Severenity
type Severenity int

const (
	UNKNOW Severenity = iota
	INFO
	WARNING
	CRITICAL
)

type ScanResult struct {
	File          *FileInfo
	Sever         Severenity
	FirstByte     int
	AfterLastByte int
	Title         string
}

type FileInfo struct {
	Path    string
	content []byte
}

func NewFileInfo() *FileInfo {
	return &FileInfo{}
}

// result bytes shared by all scanners and MUST not be change
func (this *FileInfo) GetContent() []byte {
	return this.content
}

// Return nil if nothing was finded
type Rule func(f *FileInfo) []ScanResult

var Readers sync.WaitGroup
var Scanners sync.WaitGroup
var WritersResult sync.WaitGroup

//go:generate go-bindata manulBase.xml
func Main() {
	var err error
	rules := []Rule{}

	var manulConfig = pflag.String("manul", "INTERNAL", "Path to manul config or \"INTERNAL\" for builtin rules usage or \"NONE\" for ignore manul rules")
	var help = pflag.BoolP("help", "h", false, "show this help message")
	var readersCount = pflag.Int("readers", 0, "Count of readers. Default equal to core counts/2+1")
	var scannerCount = pflag.Int("scanners", 0, "Count of parallel scanners of already readed file contents. Default equal to core count/2+1")
	pflag.Int64Var(&MAX_FILE_SIZE, "MaxFileSize", MAX_FILE_SIZE, "Files more then MaxFileSize bytes will be skip while scan.")

	pflag.Parse()

	if *help {
		printHelp()
		return
	}

	// Read manul config
	if *manulConfig != "NONE" {
		var xmlConfig ManulConfig
		var xmlBinary []byte
		if *manulConfig == "INTERNAL" {
			xmlBinary, _ = manulbaseXmlBytes()
		} else {
			xmlBinary, err = ioutil.ReadFile(*manulConfig)
			if err != nil {
				log.Printf("Can't read manul config: '%v', err: '%v'", *manulConfig, err)
			}
		}
		if xmlBinary != nil {
			err = xml.Unmarshal(xmlBinary, &xmlConfig)
			if err == nil {
				for _, manulRule := range xmlConfig.Rules {
					rule := manulRule.ToRule()
					if rule != nil {
						rules = append(rules, rule)
					}
				}
			} else {
				log.Printf("Error while parse manul config: '%v', err:'%v'", *manulConfig, err)
			}
		}
	}


	// start process
	if *readersCount == 0 {
		*readersCount = runtime.NumCPU()/2 + 1
	}
	readOrder := make(chan *FileInfo, *readersCount)
	if *scannerCount == 0 {
		*scannerCount = runtime.NumCPU()/2 + 1
	}
	scanOrder := make(chan *FileInfo, *scannerCount)
	StartReaders(readOrder, scanOrder, *readersCount)

	resultOrder := make(chan ScanResult, runtime.NumCPU())
	StartScanners(rules, scanOrder, resultOrder, *scannerCount)

	WritersResult.Add(1)
	go func() {
		PrintFileNameAndRule(resultOrder)
		WritersResult.Done()
	}()

	// Start read filenames
	if pflag.NArg() > 0 {
		ReadFilesFromArgs(readOrder, pflag.Args()...)
	} else {
		go ReadFilesFromStdIn(readOrder)
	}

	// Graceful stop
	Readers.Wait()
	log.Println("Close scan order")
	close(scanOrder)

	Scanners.Wait()
	close(resultOrder)

	WritersResult.Wait()
}

func printHelp() {
	fmt.Print(`Usage:
gnul [options] files...

or

gnul [options]
without files... - then it will be readed from stdin, a file by line until EOF

OPTIONS:
`)
	pflag.PrintDefaults()
}
