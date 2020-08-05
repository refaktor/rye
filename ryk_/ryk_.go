package ryk_

import (
	"../env"
	//	"bufio"
	"encoding/xml"
	//sb	"flag"
	"fmt"
	"os"
	//	"strings"
	//	"regexp"
	//	"net/url"
)

func ryk_xml_main() {

	//	rmap := map[string]
	//	rmap := make(map[string]interface{"PS": env.RawMap(map[string]interface{"N": env.NewBlock(env.NewTSeries([]env.Object{env.Word{1}})){}, "name"})}, 100)
	//	rmap["PS"] = rmap2

	mmap := *env.NewRawMap(rmap)

	decoder := xml.NewDecoder(os.Stdin)
	//total := 0
	var inElement string
	var inN bool
	var prn bool
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			inElement = se.Name.Local
			switch inElement {
			case "PopolnoIme", "Posta", "Ulica":
				prn = true
			}
			if inN {
				//fmt.Println(se.Attr)
			}
			if inElement == "PS" {
				inN = true
				if inN {
					fmt.Print(se.Attr[0].Value)
					fmt.Print(";")
				}
			}
			if inElement == "N" {
				if inN {
					fmt.Print(se.Attr[0].Value)
					fmt.Print(";")
					if len(se.Attr) > 1 {
						fmt.Print(se.Attr[1].Value)
						fmt.Print(";")
					}
				}
			}
		case xml.CharData:
			if prn {
				fmt.Print(string(se.Copy()))
				fmt.Print(";")
			}
		case xml.EndElement:
			inElement = se.Name.Local
			if inElement == "PS" {
				fmt.Println(inElement)
				inN = false
			}
			prn = false

		default:
		}
	}
}
