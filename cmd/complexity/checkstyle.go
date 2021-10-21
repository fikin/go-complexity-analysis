package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

type checkstyleErrorTag struct {
	XMLName  xml.Name `xml:"error"`
	Col      int      `xml:"column,attr"`
	Line     int      `xml:"line,attr"`
	Msg      string   `xml:"message,attr"`
	Severity string   `xml:"severity,attr,omitempty"`
	Source   string   `xml:"source,attr,omitempty"`
}
type checkstyleFileTag struct {
	XMLName  xml.Name `xml:"file"`
	FileName string   `xml:"name,attr"`
	Errors   []checkstyleErrorTag
}

// checkstyleTag is structure used to serialize in xml all diagnostic
type checkstyleTag struct {
	XMLName    xml.Name `xml:"checkstyle"`
	Version    string   `xml:"version,attr"`
	filesAsMap map[string]checkstyleFileTag
	Files      []checkstyleFileTag
}

func doPrintcheckstyles(data checkstyleTag) {
	for _, v := range data.filesAsMap {
		data.Files = append(data.Files, v)
	}
	output, err := xml.MarshalIndent(data, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	os.Stdout.Write(output)
}
