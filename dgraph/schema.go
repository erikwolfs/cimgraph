package dgraph

import (
	"fmt"
	"os"
	"io"
	"encoding/xml"
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
	PackageElements []PackageElement `xml:"packagedElement"`
	Attributes []Attribute `xml:"ownedAttribute"`
	Generalizations []Generalization `xml:"generalization"`
	Ends []End `xml:"ownedEnd"`
}

type Attribute struct {
	XMLName xml.Name `xml:"ownedAttribute"`
	AttrName     string   `xml:"name,attr"`
	Type string `xml:"type,attr"`
	PropertTypes PropertyType `xml:"type"`
}

type PropertyType struct {
	Type string `xml:"idref,attr"`
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

func CreateSchema(path string) error {
	fmt.Println(path)
	err := parseXMI(path)
	if err != nil {
		return err
	}
	return nil
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
}

func parseXMI(path string) error {
	file, err := os.Open(path)
    if err != nil {
        return err
    }
	fmt.Println("Successfully Opened ", path)
    defer file.Close()
	var xmi XMI
	var cim CIMProfile
	classindex := make(map[string]CIMClass)
	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = makeCharsetReader
	err = decoder.Decode(&xmi)
	if err != nil {
		return err
    }
	for i := 0; i < len(xmi.Models); i++ {
		for j := 0; j < len(xmi.Models[i].Packages); j++ {
			for k := 0; k < len(xmi.Models[i].Packages[j].PackageElements); k++ {
				err = processPackageElement(&xmi.Models[i].Packages[j].PackageElements[k], &cim, classindex)
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
	showCIM(&cim)
	return nil
}

func processPackageElement(element *PackageElement, cim *CIMProfile, classindex map[string]CIMClass) error {
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
		for i := 0; i < len(element.Ends); i++ {
			err := processEnd(&element.Ends[i])
			if err != nil {
				return err
			}
		}
		cim.Classes = append(cim.Classes, class)
		classindex[element.ID] = class
	}
	for i := 0; i < len(element.PackageElements); i++ {
		err := processPackageElement(&element.PackageElements[i], cim, classindex)
		if err != nil {
			return err
		}
	}
	return nil
}

func processAttribute(attr *Attribute, class *CIMClass) error {
	if attr.Type == "uml:Property" {
		prop := CIMProperty{Name: class.Name + "." + attr.AttrName, Object: attr.PropertTypes.Type}
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

func processEnd(end *End) error {
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
					newproperty := CIMProperty{Name: refclass.Properties[j].Name, Object: refclass.Properties[j].Object}
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

func showCIM(cim *CIMProfile) {
	for _, v := range cim.Classes {
		fmt.Println("Class: ", v.Name)
		for _, w := range v.InheritsFrom {
			fmt.Println(".  Generic:", w.Name)
		}
		for _, x := range v.Properties {
		 	fmt.Println(".  Property:", x.Name, " Type: ", x.Type)
		}
	}
}
