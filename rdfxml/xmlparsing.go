package rdfxml

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"context"
	"strings"
	"golang.org/x/text/encoding/charmap"
	db "cimgraph/postgres"
)

type RDFDataSet struct {
	XMLName xml.Name
	About string `xml:"about,attr"`
	Attributes []RDFDSAttribute `xml:",any"`
}

type RDFDSAttribute struct {
	XMLName xml.Name
	Recource string `xml:"resource,attr"`
	Value string `xml:",chardata"`
}

type RDFSubject struct {
	XMLName xml.Name
	ID string `xml:"ID,attr"`
	Predicates []RDFPredicate `xml:",any"`
}

type RDFPredicate struct {
	XMLName xml.Name 
	Recource string `xml:"resource,attr"`
	Value string `xml:",chardata"`
}

func ImportRDFtoDB(config *db.Config) error {
	fmt.Println("importing from: ", config.ImportPath , "into PostgresSQL server:", config.DBName, "on:", config.Host )
	var datasetid int64 = 0
	file, err := os.Open(config.ImportPath)
    if err != nil {
        return fmt.Errorf("error opening rdf file: %v", err)
    }
	fmt.Println("Successfully Opened ", config.ImportPath)
    defer file.Close()
	ctx := context.Background()
	pgcon, err := db.NewConnection(config, ctx)
 	if err != nil {
 		return err
 	}
 	pgtx, err := db.NewTransaction(pgcon, ctx)
 	if err != nil {
 		return err
 	}
 	err = db.SchemaSet(pgtx, ctx, config.Schema)
 	if err != nil {
 		return err
 	}
	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = makeCharsetReader
	for {
		t, tokenErr := decoder.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			return fmt.Errorf("error while decoding token: %v", err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Space != "RDF" {
				if t.Name.Space == "http://www.w3.org/1999/02/22-rdf-syntax-ns#" {
					continue
				} else if t.Name.Space == "http://iec.ch/TC57/61970-552/ModelDescription/1#" {
					var dataset RDFDataSet
					if err := decoder.DecodeElement(&dataset, &t); err != nil {
						return fmt.Errorf("error parsing dataset definition %v", err)
					}
					datasetid, err = writedataset(pgtx, &dataset, ctx)
					if err != nil {
						return err
					}
				} else {
					var subject RDFSubject
					if err := decoder.DecodeElement(&subject, &t); err != nil {
						return fmt.Errorf("error parsing tdf dubject: %v", err)
					}
					if datasetid == 0 {
						return fmt.Errorf("no dataset defined when writing subject: %s", subject)
					}
					err = writesubject(pgtx, datasetid , &subject, ctx)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	err = db.CommitTransaction(pgtx, ctx)
 	if err != nil {
 		return err
 	}
	db.CloseConnection(pgcon)
	file.Close()
	return nil
}

func writedataset(tx db.PostgresTx, rdfdataset *RDFDataSet, ctx context.Context) (int64, error) {
	dbdataset := db.DataSet{TSType: rdfdataset.XMLName.Local,
							RDFAbout: rdfdataset.About,}
	err := db.InsertDataSet(tx, &dbdataset, ctx)
	if err != nil {
		return 0, err
	}
	for _,i := range rdfdataset.Attributes {
		var value string
		if i.Recource != "" {
			value = i.Recource
		} else {
			value = i.Value
		}
		dbdsattr := db.DSAttribute{ATName: i.XMLName.Local,
									ATValue: value}
		err := db.InsertDSAttribute(tx, dbdataset.ID, &dbdsattr, ctx)
		if err != nil {
			return 0, err
		}
	}
	return dbdataset.ID, nil
}

func writesubject(tx db.PostgresTx, datasetid int64 , rdfsubject *RDFSubject, ctx context.Context) (error) {
	dbsubject := db.Subject{DatasetID: datasetid,
							RDFAbout: rdfsubject.ID,
							Type: rdfsubject.XMLName.Local}
	err := db.InsertSubject(tx, &dbsubject, ctx)
	if err != nil {
		return err
	}
	for _,i := range rdfsubject.Predicates {
		dbpredicate := db.Predicate{Type: i.XMLName.Local}
		err := db.InsertPredicate(tx, dbsubject.ID, &dbpredicate, ctx)
		if err != nil {
			return err
		}
		dbobject := db.Object{ValueStr: i.Value,
								ResourceURI: i.Recource,}
		if strings.HasPrefix(dbobject.ResourceURI, "http") {
			dbobject.ValueStr = dbobject.ResourceURI
			dbobject.ResourceURI = ""
		}
		err = db.InsertObject(tx, dbpredicate.ID, &dbobject, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func Test_empty(config *db.Config) error {
	ctx := context.Background()
	pgcon, err := db.NewConnection(config, ctx)
 	if err != nil {
 		return err
 	}
	pgtx, err := db.NewTransaction(pgcon, ctx)
 	if err != nil {
 		return err
 	}
 	err = db.SchemaSet(pgtx, ctx, config.Schema)
 	if err != nil {
 		return err
 	}
	var objects []*db.EmptyObject
	var subjects []*db.SubjectByRDFID
	err = db.RetrieveAboutsWithoutResource(pgtx, &objects, ctx)
	if err != nil {
 		return err
 	}
	for _, object := range objects {
		object.ResourceURI = strings.TrimPrefix(object.ResourceURI, "#")
		err = db.RetrieveSubjectByRDFID(pgtx, &subjects, object, ctx)
		if err != nil {
			return err
		}
		if len(subjects) > 1 {
			return fmt.Errorf("more then one subject found for rdfid and active period: %v", object)
		}
		for _,subject := range subjects {
			var resource = db.Resource{ObjectID: object.ID,
									SubjectID: subject.ID,}
			err = db.InsertResource(pgtx, &resource, ctx)
			if err != nil {
				return err
			}
		}
	}
	db.CommitTransaction(pgtx, ctx)
	db.CloseConnection(pgcon)
	return nil
}

func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
    if charset == "ISO-8859-1" || charset == "windows-1252" {
        // Windows-1252 is a superset of ISO-8859-1, so should do here
        return charmap.Windows1252.NewDecoder().Reader(input), nil
    }
    return nil, fmt.Errorf("unknown charset: %s", charset)
}