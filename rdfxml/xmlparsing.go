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
	"cimgraph/common"
)

//RDF Message Header Object
type RDFDataSet struct {
	XMLName xml.Name
	About string `xml:"about,attr"`
	Attributes []RDFDSAttribute `xml:",any"`
}

//RDF Message Header Attribute Object
type RDFDSAttribute struct {
	XMLName xml.Name
	Recource string `xml:"resource,attr"`
	Value string `xml:",chardata"`
}

//RDF Subject Object
type RDFSubject struct {
	XMLName xml.Name
	ID string `xml:"ID,attr"`
	Predicates []RDFPredicate `xml:",any"`
}

//RDF Predicate Object including RDF Object as value or resource
type RDFPredicate struct {
	XMLName xml.Name 
	Recource string `xml:"resource,attr"`
	Value string `xml:",chardata"`
}

//Main function called when importing a RDF XML message
func ImportRDFtoDB(config *db.Config) error {
	fmt.Println("importing from: ", config.ImportPath , "into PostgresSQL server:", config.DBName, "on:", config.Host )
	//varibel to store the DB ID given on the whole dataset when writing it to the db
	var datasetid int64 = 0
	//Open de message from file
	file, err := os.Open(config.ImportPath)
    if err != nil {
        return fmt.Errorf("error opening rdf file: %v", err)
    }
	fmt.Println("Successfully Opened ", config.ImportPath)
    defer file.Close()
	//context can be used to define timeouts when communicating with the DB
	ctx := context.Background()
	//Create a new connectionpool
	pgpool, err := db.NewConnectionPool(config, ctx)
 	if err != nil {
 		return err
 	}
	//Create a new transaction to make sure nothing is written when error ocures
 	pgtx, err := pgpool.NewTransaction(ctx)
 	if err != nil {
 		return err
 	}
	//Select the correct DB schema
 	err = pgtx.SetSchema(ctx, config.Schema)
 	if err != nil {
 		return err
 	}
	//Start parsing the XML in chunks
	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = makeCharsetReader
	t := common.CurrentTime()
	valset := &ValSet
	valresult := &ValResultSet
	for {
		t, tokenErr := decoder.Token()
		if tokenErr != nil {
			//This stops the parsing when the end of file is reached or when error during parsing
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
					//Parse the Message Header
					if err := decoder.DecodeElement(&dataset, &t); err != nil {
						return fmt.Errorf("error parsing dataset definition %v", err)
					}
					//retrieve profiles from the dataset
					dbcon, err := pgpool.NewConnection(ctx)
					if err != nil {
						return err
					}
					profiles, err := findProfilesToValidate(dbcon, &dataset, config.Schema, ctx)
					if err != nil {
						return err
					}
					dbcon.Release()
					if len(profiles) == 0 {
						return fmt.Errorf("no RDFS found to validate this profile, import ended")
					}
					valset, err = RetrieveValSetsFromDB(pgpool, profiles, ctx, config.Schema)
					if err != nil {
						return err
					}
					fmt.Println(valset)
					//write the dataset including attributes to the database
					datasetid, err = writedataset(pgtx, &dataset, ctx)
					if err != nil {
						return err
					}
					fmt.Println("dataset records written to the db")
				} else {
					var subject RDFSubject
					//Parse a subject
					if err := decoder.DecodeElement(&subject, &t); err != nil {
						return fmt.Errorf("error parsing rdf subject: %v", err)
					}
					//Dont save subject unless they are linked to a dataset, this means header has to be on top of file
					if datasetid == 0 {
						return fmt.Errorf("no dataset defined when writing subject: %s", subject)
					}
					//validate a subject
					err = ValidateRDFSubject(&subject, valset, valresult)
					for line := range *valresult {
						fmt.Println(line)
					}
					if err != nil {
						return err
					}
					//Write a subject to the DB
					err = writesubject(pgtx, datasetid , &subject, ctx)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	fmt.Println(valresult)
	//Close the RDF message file
	fmt.Println("triples written to the DB")
	file.Close()
	common.MeasureTime("write triples to DB", t)
	t = common.CurrentTime()
	//Retrieve subject subject relationships and write them to the DB using the DB IDs
	err = writerelationships(pgtx, ctx)
	 	if err != nil {
 		return err
 	}
	fmt.Println("relationships written to the DB")
	common.MeasureTime("write relationships to DB", t)
	//Commit the transaction
	err = pgtx.Commit(ctx)
 	if err != nil {
 		return err
 	}
	pgpool.Close()
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

func writesubject(tx db.PostgresTx, datasetid int64 , rdfsubject *RDFSubject, ctx context.Context) error {
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

func writerelationships(tx db.PostgresTx, ctx context.Context) error {
	var objects []*db.EmptyObject
	var subjects []*db.SubjectByRDFID
	err := db.RetrieveAboutsWithoutResource(tx, &objects, ctx)
	if err != nil {
 		return err
 	}
	for _, object := range objects {
		object.ResourceURI = strings.TrimPrefix(object.ResourceURI, "#")
		err = db.RetrieveSubjectByRDFID(tx, &subjects, object, ctx)
		if err != nil {
			return err
		}
		if len(subjects) > 1 {
			return fmt.Errorf("more then one subject found for rdfid and active period: %v", object)
		}
		for _,subject := range subjects {
			var resource = db.Resource{ObjectID: object.ID,
									SubjectID: subject.ID,}
			err = db.InsertResource(tx, &resource, ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func findProfilesToValidate(dbcon *db.PostgresConn, dataset *RDFDataSet, schema string, ctx context.Context) (map[string]int64, error) {
	profiles := make(map[string]int64)
	for _, attribute := range dataset.Attributes {
		if attribute.XMLName.Local == "Model.profile" {
			id, err := db.SearchProfile(dbcon, attribute.Value, schema, ctx)
			if err != nil {
				return nil, err
			}
			if id > 0 {
				profiles[attribute.Value] = id
			} else {
				fmt.Println("RDFS for profile " + attribute.Value + " not found")
			}
		}
	}
	return profiles, nil
}

func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
    if charset == "ISO-8859-1" || charset == "windows-1252" {
        // Windows-1252 is a superset of ISO-8859-1, so should do here
        return charmap.Windows1252.NewDecoder().Reader(input), nil
    }
    return nil, fmt.Errorf("unknown charset: %s", charset)
}