package gnul

import (
	"bytes"
	"encoding/xml"
	"log"
	"regexp"
)

type ManulRule struct {
	Id      string `xml:"id,attr"`
	Format  string `xml:"format,attr"`
	ChildId string `xml:"child_id,attr"`
	Sever   string `xml:"sever,attr"`
	Title   string `xml:"title,attr"`
	Content string `xml:",innerxml`
}

type ManulConfig struct {
	XMLName xml.Name    `xml:"database"`
	Rules   []ManulRule `xml:"signature"`
}

func (this ManulRule) ToRule() Rule {
	sever := UNKNOW
	switch this.Sever {
	case "c":
		sever = CRITICAL
	case "w":
		sever = WARNING
	case "i":
		sever = INFO
	}
	switch this.Format {
	case "const":
		fu := func(f *FileInfo) (res []ScanResult) {
			startByte := 0
			for pos := 0; pos != -1; pos = bytes.Index(f.Content[startByte:], []byte(this.Content)) {
				res = append(res, ScanResult{
					File:          f,
					Sever:         sever,
					FirstByte:     startByte + pos,
					AfterLastByte: startByte + pos + len(this.Content),
					Title:         "manul" + this.Id + " " + this.Title,
				})
			}
			return
		}
		return fu
	case "re":
		re, err := regexp.Compile(this.Content)
		if err != nil {
			log.Printf("Can't parse manul regexp, id: '%v', value='%v'\n", this.Id, this.Content)
			return nil
		}
		return func(f *FileInfo) (res []ScanResult) {
			resRanges := re.FindAllIndex(f.Content, -1)
			if resRanges == nil {
				return nil
			}
			res = make([]ScanResult, len(resRanges), 0)
			for _, resRange := range resRanges {
				res = append(res, ScanResult{
					File:          f,
					Sever:         sever,
					FirstByte:     resRange[0],
					AfterLastByte: resRange[1],
					Title:         "manul" + this.Id + " " + this.Title,
				})
			}
			return res
		}
	default:
		log.Printf("Unknow format of manul rule id: '%v', format: '%v'\n", this.Id, this.Format)
		return nil
	}
}
