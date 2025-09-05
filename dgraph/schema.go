package dgraph

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"context"

	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"golang.org/x/text/encoding/charmap"
)

type XMI struct {
    XMLName xml.Name  `xml:"XMI"`
	Key     string   `xml:"key,attr"`
    Models   []Model `xml:"Model"`
}

type Model struct {
	XMLName xml.Name `xml:"Model"`
	Type     string   `xml:"type,attr"`
	ModelName	string `xml:"name,attr"`
	Packages []Package `xml:"packagedElement"`
}

type Package struct {
	XMLName xml.Name `xml:"packagedElement"`
	PackageName     string   `xml:"name,attr"`
	Type string `xml:"type,attr"`
	PackageElements []PackageElement `xml:"packagedElement"`
	
}

type PackageElement struct {
	XMLName xml.Name `xml:"packagedElement"`
	ID string `xml:"id,attr"`
	PackageElName     string   `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Abstract bool `xml:"isAbstract,attr"`
	PackageElements []PackageElement `xml:"packagedElement"`
	Attributes []Attribute `xml:"ownedAttribute"`
	Generalizations []Generalization `xml:"generalization"`
	Ends []End `xml:"ownedEnd"`
	Literals []Literal `xml:"ownedLiteral"`
}

type Attribute struct {
	XMLName xml.Name `xml:"ownedAttribute"`
	AttrName     string   `xml:"name,attr"`
	Type string `xml:"type,attr"`
	PropertTypes PropertyType `xml:"type"`
	Lower PropertyLower `xml:"lowerValue"`
	Upper PropertyUpper `xml:"upperValue"`
}

type PropertyType struct {
	Type string `xml:"idref,attr"`
}

type PropertyLower struct {
	Value string `xml:"value,attr"`
}

type PropertyUpper struct {
	Value string `xml:"value,attr"`
}

type Generalization struct {
	XMLName xml.Name `xml:"generalization"`
	Type     string   `xml:"type,attr"`
	General string `xml:"general,attr"`
}

type End struct {
	XMLName xml.Name `xml:"ownedEnd"`
	EndName     string   `xml:"name,attr"`
	Association string `xml:"association,attr"`
}

type Literal struct {
	XMLName xml.Name `xml:"ownedLiteral"`
	ID string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type CIMProfile struct {
	Classes []CIMClass
}

type CIMClass struct {
	Name string
	ID string
	InheritsFrom []CIMInheritance
	Properties []CIMProperty
}

type CIMInheritance struct {
	ID string
	Name string
}

type CIMProperty struct {
	Object string
	Type string
	Name string
	Lower string
	Upper string
}

type CIMEnum struct {
	ID string
	Name string
	Literals []CIMEnumLiteral
}

type CIMEnumLiteral struct {
	ID string
	Name string
}

type Schema struct {
	Predicates map[string]SchemaPredicate
	Nodes map[string]SchemaNode
}

type SchemaPredicate struct {
	Name string
	Type string
	Index bool
}

type SchemaNode struct {
	Name string
	Predicates map[string]SchemaPredicate
}

func CreateSchema(path string) error {
	schema := Schema{}
	fmt.Println(path)
	err := parseXMI(path, &schema)
	if err != nil {
		return err
	}
	writeSchema(schema)
	con, err := newClient()
	if err != nil {
		return err
	}
	err = dropAllData(con)
	if err != nil {
		return err
	}
	err = writePredicates(con, &schema)
	if err != nil {
		return err
	}
	err = writeNodes(con, &schema)
	if err != nil {
		return err
	}
	con.Close()
	return nil
}

func writePredicates(conn *dgo.Dgraph, schema *Schema) error {
	for _,i := range schema.Predicates {
		err := addPredicate(conn, &i)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeNodes(conn *dgo.Dgraph, schema *Schema) error {
	for _,i := range schema.Nodes {
		err := addNode(conn, &i)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseXMI(path string, schema *Schema) error {
	file, err := os.Open(path)
    if err != nil {
        return err
    }
	fmt.Println("Successfully Opened ", path)
    defer file.Close()
	var xmi XMI
	var cim CIMProfile
	classindex := make(map[string]CIMClass)
	enumIndex := make(map[string]CIMEnum)
	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = makeCharsetReader
	err = decoder.Decode(&xmi)
	if err != nil {
		return err
    }
	for i := 0; i < len(xmi.Models); i++ {
		for j := 0; j < len(xmi.Models[i].Packages); j++ {
			for k := 0; k < len(xmi.Models[i].Packages[j].PackageElements); k++ {
				err = processPackageElement(&xmi.Models[i].Packages[j].PackageElements[k], &cim, classindex, enumIndex)
				if err != nil {
					return err
				}
			}
		}
    }
	err = searchGenerics(&cim, classindex)
	if err != nil {
		return err
	}
	err = addPropertyTypes(&cim, classindex)
	if err != nil {
		return err
	}
	err = SearchEnums(&cim, enumIndex)
	{
		if err != nil {
			return err
		}
	}
	//showCIM(&cim)
	createSchemaData(schema, &cim)
	writeCIM(&cim)
	return nil
}

func processPackageElement(element *PackageElement, cim *CIMProfile, classindex map[string]CIMClass, enumindex map[string]CIMEnum) error {
	if element.Type == "uml:Class" && element.PackageElName != "" {
		class := CIMClass{Name: element.PackageElName, ID: element.ID}
		for i := 0; i < len(element.Attributes); i++ {
			err := processAttribute(&element.Attributes[i], &class)
			if err != nil {
				return err
			}
		}
		for i := 0; i < len(element.Generalizations); i++ {
			err := processGeneralization(&element.Generalizations[i], &class)
			if err != nil {
				return err
			}
		}
		// for i := 0; i < len(element.Ends); i++ {
		// 	err := processEnd(&element.Ends[i])
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		if !element.Abstract {
			cim.Classes = append(cim.Classes, class)
		}
		classindex[element.ID] = class
	}
	if element.Type == "uml:Enumeration" && element.PackageElName != "" {
		enum := CIMEnum{Name: element.PackageElName, ID: element.ID}
		for i := 0; i < len(element.Literals); i++ {
			err := processEnum(&element.Literals[i], &enum)
			if err != nil {
				return err
			}
		}
		enumindex[element.ID] = enum
	}
	for i := 0; i < len(element.PackageElements); i++ {
		err := processPackageElement(&element.PackageElements[i], cim, classindex, enumindex)
		if err != nil {
			return err
		}
	}
	return nil
}

func processAttribute(attr *Attribute, class *CIMClass) error {
	if attr.Type == "uml:Property" {
		prop := CIMProperty{Name: class.Name + "." + attr.AttrName, 
							Object: attr.PropertTypes.Type,
							Lower: attr.Lower.Value,
							Upper: attr.Upper.Value,}
		class.Properties = append(class.Properties, prop)
	}
	return nil
}

func processGeneralization(generic *Generalization, class *CIMClass) error {
	if generic.General != "" {
		inherit := CIMInheritance{ID: generic.General}
		class.InheritsFrom = append(class.InheritsFrom, inherit)
	}
	return nil
}

// func processEnd(end *End) error {
// 	return nil
// }

func processEnum(literal *Literal, enum *CIMEnum) error {
	lit := CIMEnumLiteral{ID: literal.ID, Name: literal.Name}
	enum.Literals = append(enum.Literals, lit)
	return nil
}

func searchGenerics(cim *CIMProfile, index map[string]CIMClass) error {
	for i := 0; i < len(cim.Classes); i++ {
		if len(cim.Classes[i].InheritsFrom) > 0 {
			count := 0
			for {
				refclass := index[cim.Classes[i].InheritsFrom[count].ID]
				cim.Classes[i].InheritsFrom[count].Name = refclass.Name
				for j := 0; j < len(refclass.Properties); j++ {
					newproperty := CIMProperty{Name: refclass.Properties[j].Name,
												Object: refclass.Properties[j].Object,
												Lower: refclass.Properties[j].Lower,
												Upper: refclass.Properties[j].Upper,}
					cim.Classes[i].Properties = append(cim.Classes[i].Properties, newproperty)
				}
				if len(refclass.InheritsFrom) == 1 {
					newinherit := CIMInheritance{ID: refclass.InheritsFrom[0].ID}
					cim.Classes[i].InheritsFrom = append(cim.Classes[i].InheritsFrom, newinherit)
					count++
				} else {
				break
				}
			}
			
		}
	}
	return nil
}

func SearchEnums(cim *CIMProfile, index map[string]CIMEnum) error {
	for i := 0; i < len(cim.Classes); i++ {
		for j := 0; j < len(cim.Classes[i].Properties); j++ {
			if cim.Classes[i].Properties[j].Type == "" {
				refenum := index[cim.Classes[i].Properties[j].Object]
				cim.Classes[i].Properties[j].Type = refenum.Name
			}
		}
	}
	return nil
}

func addPropertyTypes(cim *CIMProfile, index map[string]CIMClass) error {
	for i := 0; i < len(cim.Classes); i++ {
		for j := 0; j < len(cim.Classes[i].Properties); j++ {
			refclass := index[cim.Classes[i].Properties[j].Object]
			cim.Classes[i].Properties[j].Type = refclass.Name
		}
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

// func showCIM(cim *CIMProfile) {
// 	for _, v := range cim.Classes {
// 		fmt.Println("Class: ", v.Name)
// 		for _, w := range v.InheritsFrom {
// 			fmt.Println(".  Generic:", w.Name)
// 		}
// 		for _, x := range v.Properties {
// 		 	fmt.Println(".  Property:", x.Name, " Type: ", x.Type)
// 		}
// 	}
// }

func writeCIM(cim *CIMProfile) {
	file, err := os.Create("./data/output.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(file)
	for _, x := range cim.Classes {
		_, err := w.WriteString("Class: " + x.Name + "\n")
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, y := range x.Properties {
			_, err := w.WriteString(".  Property: " + y.Name + " Type: " + y.Type + " Upper " + y.Upper +"\n")
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	w.Flush()
}

func createSchemaData(schema *Schema, cim *CIMProfile) error {
	var predtype string
	var node SchemaNode
	var predicate SchemaPredicate
	var existing bool
	rdfpredicate := SchemaPredicate{Name: "rdf.about", Type: "string"}
	schema.Nodes = make(map[string]SchemaNode)
	schema.Predicates = make(map[string]SchemaPredicate)
	schema.Predicates["rdf.about"] = rdfpredicate
	for _, i := range cim.Classes {
		if len(i.Properties) > 1 {
			node, existing = schema.Nodes[i.Name]
			if !existing {
				node = SchemaNode{Name: i.Name}
				node.Predicates = make(map[string]SchemaPredicate)
				node.Predicates["rdf.about"] = rdfpredicate
				schema.Nodes[i.Name] = node
				node = schema.Nodes[i.Name]
			}
			for _, j := range i.Properties {
				switch j.Type {
				case "String":
					predtype = "string"
				case "Float":
					predtype = "float"
				case "Simple_Float":
					predtype = "float"
				case "Boolean":
					predtype = "bool"
				case "Integer":
					predtype = "int"
				case "DateTime":
					predtype = "dateTime"
				case "Date":
					predtype = "dateTime"
				default:
					predtype = "uid"
				}
				if j.Upper == "*" {
					predtype = "[" + predtype + "]"
				}
				_, existing = schema.Predicates[j.Name]
				if !existing {
					predicate = SchemaPredicate{Name: j.Name, Type: predtype}
					schema.Predicates[j.Name] = predicate
				}
				_, existing = node.Predicates[j.Name]
				if !existing {
					predicate = SchemaPredicate{Name: j.Name}
					node.Predicates[j.Name] = predicate
				}
			}	
		}
	}
	println("created predicates")
	for _, i := range schema.Predicates {
		fmt.Println("<" + i.Name + ">: " + i.Type + " .")
	}
	println("created nodes")
	for _, i := range schema.Nodes {
		fmt.Println ("type <" + i.Name + "> {")
		for _, j := range i.Predicates {
			fmt.Println("  " + j.Name)
		}
		fmt.Println("}")
	}
	return nil
}

func writeSchema(schema Schema) {
	file, err := os.Create("./data/schema.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(file)
	for _, i := range schema.Predicates {
		_, err := w.WriteString("<" + i.Name + ">: " + i.Type + " .\n")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	for _, i := range schema.Nodes {
		_, err := w.WriteString("type <" + i.Name + "> {\n")
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, j := range i.Predicates {
			_, err := w.WriteString("  " + j.Name + "\n")
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		_, err = w.WriteString("}\n")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	w.Flush()
}

func addPredicate(conn *dgo.Dgraph, predicate *SchemaPredicate) error {
	err := conn.Alter(context.Background(), &api.Operation{
    Schema: "<" + predicate.Name + ">: " + predicate.Type + " .",
  	})
	if err != nil {
		return err
	}
	return nil
}

func addNode(conn *dgo.Dgraph, node *SchemaNode) error {
	schemastr := "type <" + node.Name + "> {\n"
	for _, j := range node.Predicates {
		schemastr = schemastr + j.Name + "\n"
	}
	schemastr = schemastr + "}\n"
	err := conn.Alter(context.Background(), &api.Operation{
    Schema: schemastr,
  	})
	if err != nil {
		return err
	}
	println("node " + node.Name + " added to schema")
	return nil
}
