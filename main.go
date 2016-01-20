package gnul
import (
	"github.com/rekby/pflag"
	"encoding/xml"
	"log"
	"io/ioutil"
	"fmt"
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
	Content []byte
}

// Return nil if nothing was finded
type Rule func(f *FileInfo) []ScanResult

var manulConfig = pflag.String("manul","INTERNAL","Path to manul config or \"INTERNAL\" for builtin rules usage or \"NONE\" for ignore manul rules")
var help = pflag.BoolP("help", "h", false, "show this help message")

//go:generate go-bindata manulBase.xml
func Main() {
	var err error
	rules := []Rule{}

	pflag.Parse()

	if *help {
		printHelp()
		return
	}

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


}

func printHelp(){
	fmt.Print(`Usage:
gnul [options] files...

or

gnul [options]
without files... - then it will be readed from stdin, a file by line until EOF

OPTIONS:
`)
	pflag.PrintDefaults()
}