package dgraph

import (
	"os"
	"fmt"
	"encoding/xml"
	"net/url"
	"strings"
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

type DataNode struct {
	RDFAbout string
	DGraphType string
	Predicates []DataPredicate

}

type DataPredicate struct {
	Predicate string
	Object string
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
	var object string
	var muttxt string
	con, err := newConnection(config)
	if err != nil {
		return err
	}
	txn, err := newTransaction(con)
	if err != nil {
		return err
	}
	for _,i := range doc.Subjects {
		muttxt = ""
		if i.ID != "" {
			nodeid = i.ID
		} else if i.About != "" {
			nodeid = i.About
		} else {
			return fmt.Errorf("subject has no id can not parse")
		}
		//hashnr := nhash(nodeid)
		muttxt = "<_:" + nodeid + "> <rdf.about> \"" + nodeid + "\" .\n"
		muttxt = muttxt + "<_:" + nodeid + "> <dgraph.type> \"" + i.XMLName.Local + "\" .\n"
		for _,j := range i.Predicates {
			if j.Recource != "" {
				_, err := url.ParseRequestURI(j.Recource)
				if err != nil {
					object = "<" + strings.TrimPrefix(j.Recource, "#") + ">"
				} else {
					if strings.Contains(j.Recource, "http://") {
						object = "\"" + j.Recource + "\""
					} else {
						object = "\"" + j.Recource + "\""
					}
				}
				
			} else if j.Value != "" {
				object = "\"" + j.Value + "\""
			} else {
				return fmt.Errorf("object has no value or resource, can not parse")
			}
			muttxt = muttxt + "<_:" + nodeid + "> <" + j.XMLName.Local + "> " + object + " .\n"
		}
		fmt.Println(muttxt)
		err = writeMutation(txn, muttxt)
		if err != nil {
			return err
		}
	}
	err = commitTransaction(txn)
	if err != nil {
		return err
	}
	con.Close()
	return nil
}

