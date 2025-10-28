package postgres

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	//"net/url"
	"os"
	//"strings"
	//"golang.org/x/text/date"
	"golang.org/x/text/encoding/charmap"
	//"google.golang.org/genproto/googleapis/type/datetime"
)

type RDFDoc struct {
	XMLName 	xml.Name
	Subjects []RDFSubject `xml:",any"`
}

type RDFSubject struct {
	XMLName xml.Name
	About string `xml:"about,attr"`
	ID string `xml:"ID,attr"`
	Predicates []RDFPredicate `xml:",any"`
}

type RDFPredicate struct {
	XMLName xml.Name 
	Recource string `xml:"resource,attr"`
	Value string `xml:",chardata"`
}

func ImportRDF (config Config) error {
	var doc RDFDoc
	err := parseRDF(config, &doc)
	if err != nil {
		return err
	}
	err = saveRDF(config, &doc)
	if err != nil {
		return err
	}
	return nil
}

func parseRDF(config Config, doc *RDFDoc) error {
	//var doc2 Node
	file, err := os.Open(config.ImportPath)
    if err != nil {
        return err
    }
	fmt.Println("Successfully Opened ", config.ImportPath)
    defer file.Close()
	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = makeCharsetReader
	err = decoder.Decode(&doc)
	if err != nil {
		return err
    }
	return nil
}

func saveRDF(config Config, doc *RDFDoc) error {
	var nodeid string
	ctx := context.Background()
	postgres, err := newConnection(config, ctx)
	if err != nil {
		return err
	}
	tx, err := newTransaction(postgres, ctx)
	if err != nil {
		return err
	}
	err = schemaSet(tx, ctx, "cimgraph")
	if err != nil {
		return err
	}
	for _,i := range doc.Subjects {
		if i.ID != "" {
			nodeid = i.ID
		} else if i.About != "" {
			nodeid = i.About
		} else {
			return fmt.Errorf("subject has no id can not parse")
		}
		subject := Subject{RDFAbout: nodeid, Type: i.XMLName.Local}
		err = InsertSubject(tx, &subject, ctx)
		if err != nil {
			return err
		}
	}
	err = commitTransaction(tx, ctx)
	if err != nil {
		return err
	}
	return nil
}

func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
    if charset == "ISO-8859-1" || charset == "windows-1252" {
        // Windows-1252 is a superset of ISO-8859-1, so should do here
        return charmap.Windows1252.NewDecoder().Reader(input), nil
    }
    return nil, fmt.Errorf("unknown charset: %s", charset)
}

// for _,j := range i.Predicates {
		// 	predicate := Predicate{Type: j.XMLName.Local}
		// 	if j.Recource != "" {
		// 		_, err := url.ParseRequestURI(j.Recource)
		// 		if err != nil {
		// 			object := Object{ValueStr: strings.TrimPrefix(j.Recource, "#")}
		// 		} else {
		// 			if strings.Contains(j.Recource, "http://") {
		// 				object = "\"" + j.Recource + "\""
		// 			} else {
		// 				object = "\"" + j.Recource + "\""
		// 			}
		// 		}
				
		// 	} else if j.Value != "" {
		// 		object = "\"" + j.Value + "\""
		// 	} else {
		// 		return fmt.Errorf("object has no value or resource, can not parse")
		// 	}
		// 	muttxt = muttxt + "<_:" + nodeid + "> <" + j.XMLName.Local + "> " + object + " .\n"
//}


