package rdfxml

import (
	"encoding/xml"
	"fmt"
	"os"
	"io"
	"net/url"
	"strings"
	"strconv"
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

type RDFSClass struct {
	Name string
	Label string
	Comment string
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
	var classes []RDFSClass 
	var attributes []RDFSAttribute
	var associations []RDFSAssociation
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
								classes = append(classes, class)
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
								fmt.Println(association.MultiMin, association.MultiMax)
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
