package rdfxml

import (
	"encoding/xml"
	"fmt"
	"os"
	"io"
	"net/url"
	"strings"
	"strconv"
	"context"
	db "cimgraph/postgres"
)

type RDFSDescription struct {
	XMLName xml.Name
	RDFAbout string `xml:"about,attr"`
	Definitions []RDFSDefinition`xml:",any"`
}

type RDFSDefinition struct {
	XMLName xml.Name
	Lang string `xml:"lang,attr"`
	Recource string `xml:"resource,attr"`
	DataType string `xml:"datatype,attr"`
	Value string `xml:",chardata"`
}

type RDFSHeader struct {
	Identifier string
	Version string
	Items []RDFSItem
}

type RDFSItem struct {
	Name string
	Value string
}

type RDFSClass struct {
	ID int64
	Name string
	Label string
	Comment string
	SubClassOf string
	Attributes map[string]RDFSAttribute
	Associations map[string]RDFSAssociation
}

type RDFSBaseClass struct {
	Name string
	SubClassOf string
}

type RDFSAttribute struct {
	Name string
	Label string
	Domain string
	DataType string
	Mandatory bool
	Comment string
}

type RDFSAssociation struct {
	Name string
	Label string
	Domain string
	DataType string
	MultiMin int
	MultiMax int
	Comment string
	Range string
	InverseRoleName string
	Used bool
}

type RDFSEnumeration struct {
	Name string
	Label string
	Comment string
}

type RDFSEnumValue struct {
	Name string
	Label string
	Comment string
	Enumerator string
}

func ImportRDFStoDB(config *db.Config) error {
	fmt.Println("importing RDFschema from: ", config.ImportPath , "into PostgresSQL server:", config.DBName, "on:", config.Host )
	file, err := os.Open(config.ImportPath)
	var enums []RDFSEnumeration
	var enumvalues []RDFSEnumValue
	classes := make(map[string]*RDFSClass)
	var attributes []RDFSAttribute
	var associations []RDFSAssociation
	var rdfsheader RDFSHeader
    if err != nil {
        return fmt.Errorf("error opening rdfs file: %v", err)
    }
	fmt.Println("Successfully Opened ", config.ImportPath)
    defer file.Close()
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
				if t.Name.Local == "Description" {
					var description RDFSDescription
					if err := decoder.DecodeElement(&description, &t); err != nil {
						return fmt.Errorf("error parsing description definition %v", err)
					}
					var destype string
					content := make(map[string]string)
					for _, definition := range description.Definitions {
						switch definition.XMLName.Local {
							case "type":
								destype = definition.Recource
							default:
								if definition.Recource != "" {
									content[definition.XMLName.Local] = definition.Recource
								} else {
									content[definition.XMLName.Local] = definition.Value
								}
						}
					}
					destype = urlfragment(destype)
					switch destype {
						case "Class":
							if urlfragment(contains("stereotype", content)) == "enumeration" {
								enum := RDFSEnumeration{Name: description.RDFAbout,
														Label: contains("label", content),
														Comment: contains("comment", content),}
								enums = append(enums, enum)
							} else {
								class := RDFSClass{Name: description.RDFAbout,
													Label: contains("label", content),
													Comment: contains("comment", content),
													SubClassOf: contains("subClassOf", content),}
								class.Attributes = make(map[string]RDFSAttribute)
								class.Associations = make(map[string]RDFSAssociation)
								classes[description.RDFAbout] = &class
							}
						case "Property":
							if urlfragment(contains("stereotype", content)) == "attribute" {
								attribute := RDFSAttribute{Name: description.RDFAbout,
															Label: contains("label", content),
															Domain: contains("domain", content),
															DataType: contains("dataType", content),
															Mandatory: mandatory(urlfragment(contains("multiplicity",content))),
															Comment: contains("comment", content),}
								attributes = append(attributes, attribute)
							} else {
								association := RDFSAssociation{Name: description.RDFAbout,
																Label: contains("label", content),
																Domain: contains("domain", content),
																DataType: contains("dataType", content),
																MultiMin: multimin(urlfragment(contains("multiplicity",content))),
																MultiMax: multimax(urlfragment(contains("multiplicity",content))),
																Comment: contains("comment", content),
																Range: contains("range", content),
																InverseRoleName: contains("inverseRoleName", content),
																Used: strings.ToLower(contains("AssociationUsed", content)) == "yes",}
								associations = append(associations, association)
							}
						case "Ontology":
							rdfsheader = RDFSHeader{Identifier: contains("identifier", content),
													Version: contains("versionIRI", content),
														}
							for k,v := range content {
								rdfsitem := RDFSItem{Name: k,
													Value: v,}
								rdfsheader.Items = append(rdfsheader.Items, rdfsitem)
							}
						case "ClassCategory":
							for k,v := range content {
								rdfsitem := RDFSItem{Name: k,
													Value: v,}
								rdfsheader.Items = append(rdfsheader.Items, rdfsitem)
							}
						default:
							if contains("stereotype", content) == "enum" {
								enumvalue := RDFSEnumValue{Name: description.RDFAbout,
															Label: contains("label", content),
															Comment: contains("comment", content),
															Enumerator: destype,}
								enumvalues = append(enumvalues, enumvalue)
							}
					}
				}
		}
	}
	file.Close()
	err = mapattributes(classes, attributes)
	if err != nil {
		return err
	}
	err = mapassosiations(classes, associations)
	if err != nil {
		return err
	}
	err = mapinheritance(classes)
	if err != nil {
		return err
	}
	ctx := context.Background()
	pgcon, err := db.NewConnection(config, ctx)
 	if err != nil {
 		return err
 	}
 	pgtx, err := db.NewTransaction(pgcon, ctx)
 	if err != nil {
 		return err
 	}
 	err = db.SchemaSetonTx(pgtx, ctx, config.Schema)
 	if err != nil {
 		return err
 	}
	headerid, err := writerdfsheader(pgtx, &rdfsheader, ctx)
	if err != nil {
		return err
	}
	err = writeclasses(pgtx, classes, headerid, ctx)
	if err != nil {
		return err
	}
	err = db.CommitTransaction(pgtx, ctx)
 	if err != nil {
 		return err
 	}
	fmt.Println("rdf record with id:", headerid, "written")
	db.CloseConnection(pgcon)
	return nil
}

func writerdfsheader(tx db.PostgresTx, header *RDFSHeader, ctx context.Context) (int64, error) {
	dbheader := db.RDFS{RDFType: header.Identifier,
						Version: header.Version}
	err := db.InsertRDFS(tx, &dbheader, ctx)
	if err != nil {
		return 0, err
	}
	for _,i := range header.Items {
		dbitem := db.RDFSItem{Name: i.Name,
								Value: i.Value,
								RDF_ID: dbheader.ID,}
		err = db.InsertRDFSItem(tx, &dbitem, ctx)
		if err != nil {
			return 0, err
		}
	}
	return dbheader.ID, nil
}

func writeclasses(tx db.PostgresTx, classes map[string]*RDFSClass, rdfid int64, ctx context.Context) error {
	for _, i := range classes {
		dbclass := db.Class{Name: i.Name,
							Label: i.Label,
							Comment: i.Comment,
							RDFS_ID: rdfid,}
		err := db.InsertClass(tx, &dbclass, ctx)
		if err != nil {
			return err
		}
		i.ID = dbclass.ID
	}
	return nil
}

func mapattributes(classes map[string]*RDFSClass, attributes []RDFSAttribute) error {
	for _, attribute := range attributes {
		if class, ok := classes[attribute.Domain]; ok {
			class.Attributes[attribute.Name] = attribute
		} else {
			return fmt.Errorf("RDFS Attribute %v with domain %v not linked to a class", attribute.Name, attribute.Domain)
		}
	}
	return nil
}

func mapassosiations(classes map[string]*RDFSClass, associations []RDFSAssociation) error {
	for _, association := range associations {
		if class, ok := classes[association.Domain]; ok {
			class.Associations[association.Name] = association
		} else {
			return fmt.Errorf("RDFS Association %v with domain %v not linked to a class", association.Name, association.Domain)
		}
	}
	return nil
}

func mapinheritance(classes map[string]*RDFSClass) error {
	for _, class := range classes {
		if class.SubClassOf != "" {
			var baseclasses []RDFSBaseClass
			subclass := class
			nomorebaseclasses:
			for {
				baseclass := RDFSBaseClass{Name: subclass.SubClassOf}
				if classes[subclass.SubClassOf].SubClassOf != "" {
					baseclass.SubClassOf = classes[subclass.SubClassOf].SubClassOf
				}
				baseclasses = append(baseclasses, baseclass)
				if baseclass.SubClassOf != "" {
					subclass = classes[baseclass.Name]
				} else {
					break nomorebaseclasses
				}
			}
			for _, baseclass := range baseclasses {
				for _, attribute := range classes[baseclass.Name].Attributes {
					if _, ok := class.Attributes[attribute.Name]; ok {
						continue
					} else {
						class.Attributes[attribute.Name] = attribute
					}
				} 
				for _, association := range classes[baseclass.Name].Associations {
					if _, ok := class.Associations[association.Name]; ok {
						continue
					} else {
						class.Associations[association.Name] = association
					}
				} 
			} 
		}
	}
	return nil
}

func urlfragment(urlstr string) (string) {
	u, err := url.Parse(urlstr)
    if err != nil {
        return urlstr
    } else if u.Hostname() == "" {
		return urlstr
	}
	return u.Fragment
}

func contains(key string, content map[string]string) string {
	if val, ok := content[key]; ok {
		return val
	}
	return ""
}

func mandatory(multiplicity string) bool {
	return multiplicity == "M:1..1"
}

func multimin(multiplicity string) int {
	if strings.HasPrefix(multiplicity, "M:1") {
		return 1
	} else {
		return 0
	}
}

func multimax(multiplicity string) int {
	if strings.HasSuffix(multiplicity, "n") {
		return -1
	} else {
		i, err := strconv.Atoi(string(multiplicity[len(multiplicity)-1]))
		if err != nil {
			panic(err)
		}
		return i
	}
}
