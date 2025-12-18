package rdfxml

import (
	"cimgraph/common"
	db "cimgraph/postgres"
	"context"
	"fmt"
	"strconv"
)

var ValSet map[string]ValProf
var ValResultSet []ValResult

type ValProf struct {
	Classes map[string]RDFSClass
}

type ValResult struct {
	Error string
	SubjectName string
	SubjectID string
	Profile string
	Severity string
}

type ValPredicate struct {
	predicate RDFPredicate
	count int
}

func RetrieveValSetsFromDB(pool *db.PostgresPool,
							profiles map[string]int64,
							ctx context.Context,
							schema string) (*map[string]ValProf, error) {
	t := common.CurrentTime()
	conn, err := pool.NewConnection(ctx)
	if err != nil {
		return nil, err
	}
	conn.SetSchema(schema, ctx)
	ValSet := make(map[string]ValProf)
	for key, value := range profiles {
		rdfs, err := db.RetrieveProfileRDFS(conn, value, ctx)
		if err != nil {
			return nil,  err
		}
		var valprof ValProf
		valprof.Classes = make(map[string]RDFSClass) 
		for _, class := range rdfs {
			rdfsclass := RDFSClass{ID: class.Class_id, Name: class.Class_name}
			if len(class.Attr_name) > 0 {
				attributes := make(map[string]RDFSAttribute)
				for i, name := range class.Attr_name {
					attributes[name] = RDFSAttribute{Name: name,
													Mandatory: class.Attr_man[i],}
				}
				rdfsclass.Attributes = attributes
			}
			if len(class.Ass_name) > 0 {
				associations := make(map[string]RDFSAssociation)
				for i, name := range class.Ass_name {
					associations[name] = RDFSAssociation{Name: name,
															MultiMin: class.Ass_min[i],
															MultiMax: class.Ass_max[i],
															Range: class.Ass_range[i],}
				}
				rdfsclass.Associations = associations
			}
			valprof.Classes[class.Class_name] = rdfsclass
		}
		ValSet[key] = valprof
	}
	common.MeasureTime("retrieve RDFS Schema from DB", t)
	return &ValSet, nil
}

func ValidateRDFSubject(subject *RDFSubject, profileset *map[string]ValProf, result *[]ValResult) error {
	for rdfs, profile := range *profileset {
		class, ok := profile.Classes[subject.XMLName.Local]
		if !ok {
			valresult := ValResult{Error: "Subject not found in Profile RDFS",
									SubjectName: subject.XMLName.Local,
									SubjectID: subject.ID,
									Profile: rdfs,
									Severity: "low",}
			*result = append(*result, valresult)
			continue
		} else {
			valresult, err := validateRDFSPredicates(subject, &class, rdfs)
			if err != nil {
				return err
			}
			*result = append(*result, valresult...)
		}
	}
	return nil
} 

func validateRDFSPredicates(subject *RDFSubject, class *RDFSClass, profile string) ([]ValResult, error) {
	valPredicats := make(map[string]ValPredicate)
	var valresults []ValResult
	for _, predicate := range subject.Predicates {
		valpredicate, ok := valPredicats[predicate.XMLName.Local]
		if !ok {
			//if valpredicate is not existing create it
			valPredicats[predicate.XMLName.Local] = ValPredicate{predicate: predicate, count: 1}
		} else {
			//if already exists increase count with one
			valpredicate.count++
		}
	}
	//Test if required attributes are provided
	for _, attribute := range class.Attributes {
		if attribute.Mandatory {
			_, ok := valPredicats[attribute.Name]
			if !ok {
				valresult := ValResult{Error: "mandarory predicate missing",
										SubjectName: subject.XMLName.Local,
										SubjectID: subject.ID,
										Profile: profile,
										Severity: "blocking",}
				valresults = append(valresults, valresult)
				return valresults, fmt.Errorf("blocking error while validating predicates for subject: %s", class.Name)
			}
		}
	}
	//Test for required assossiations
	for name, association := range class.Associations {
		if association.Used {
			predicate, ok := valPredicats[name]
			if !ok && association.MultiMin > 0 {
				valresult := ValResult{Error: "mandarory predicate (" + association.Name +") missing",
												SubjectName: subject.XMLName.Local,
												SubjectID: subject.ID,
												Profile: profile,
												Severity: "blocking",}
				valresults = append(valresults, valresult)
				return valresults, fmt.Errorf("blocking error while validating predicates for subject: %s", class.Name)
			}
			if predicate.count < association.MultiMin || predicate.count > association.MultiMax {
				errorstr := "wrong amount of dependencies for " + association.Name + " expected between " + strconv.Itoa(association.MultiMin) + " and " + strconv.Itoa(association.MultiMax) + " found " + strconv.Itoa(predicate.count)
				valresult := ValResult{Error: errorstr,
										SubjectName: subject.XMLName.Local,
										SubjectID: subject.ID,
										Profile: profile,
										Severity: "blocking",}
				valresults = append(valresults, valresult)
				return valresults, fmt.Errorf("blocking error while validating predicates for subject: %s", class.Name)
			}
		}
	}
	//Test if there are undefined predicates
	for _, predicate := range subject.Predicates {
		_, ok := class.Attributes[predicate.XMLName.Local]
		if !ok {
			_, ok := class.Associations[predicate.XMLName.Local]
				if !ok {
					valresult := ValResult{Error: "a predicate was found nat is not defined in de RDFSchema",
											SubjectName: subject.XMLName.Local,
											SubjectID: subject.ID,
											Profile: profile,
											Severity: "low",}
					valresults = append(valresults, valresult)
				}
		}
	}
	return valresults, nil
}

