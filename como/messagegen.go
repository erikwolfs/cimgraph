package como

import (
	db "cimgraph/postgres"
	"context"
	"encoding/json"
	"fmt"
)

type ComoMessage struct {
	Subjects []Subject
}

type Subject struct {
	Cim string
	RdfID string
	Predicates []Predicate `json:"Predicates,omitempty"`
	Children []Subject `json:"Children,omitempty"`
}

type Predicate struct {
	Cim string
	Value string `json:"Value,omitempty"`
	Resource string `json:"Resource,omitempty"`
}

func GenerateGraphBySubjectType(config *db.Config, subtype string) error {
	ctx := context.Background()
	pgcon, err := db.NewConnection(config, ctx)
 	if err != nil {
 		return err
 	}
	err = db.SchemaSetonCon(*pgcon, ctx, config.Schema)
	if err != nil {
		return err
	}
	var message = ComoMessage{}
	var recsubjs []*db.SubjectByType
	err = db.RetrieveSubjectsByType(pgcon, &subtype, &recsubjs, ctx)
	if err != nil {
		return err
	}
	for _,recsubj := range recsubjs {
		subject := Subject{Cim: recsubj.Type,
						RdfID: recsubj.RDFAbout}
		err = retrievePredicatesbySubject(pgcon, recsubj.ID, &subject, ctx)
		if err != nil {
			return err
		}
		err = retrieveChildsbySubject(pgcon, recsubj.ID, &subject, ctx)
		if err != nil {
			return err
		}
		message.Subjects = append(message.Subjects, subject)
	}
	b, err := json.MarshalIndent(message, " ", "  ")
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}

func retrievePredicatesbySubject(conn *db.Postgresdb, subid int64, subject *Subject, ctx context.Context) error {
	var recpreds []*db.PredicateBySubjectID
	err := db.RetrievePredicatesBySubjectID(conn, subid, &recpreds, ctx)
	if err != nil {
		return err
	}
	for _, recpred := range recpreds {
		predicate := Predicate{Cim: recpred.Type,
								Value: recpred.ValueStr,
								Resource: recpred.ResourceURI,}
		subject.Predicates = append(subject.Predicates, predicate)
	}
	return nil
}

func retrieveChildsbySubject(conn *db.Postgresdb, subid int64, subject *Subject, ctx context.Context) error {
	var recchilds []*db.ChildSubjectByID
	err := db.RetrieveChildsBySubjectID(conn, subid, &recchilds, ctx)
	if err != nil {
		return err
	}
	for _, recchild := range recchilds {
		child := Subject{Cim: recchild.Type,
							RdfID: recchild.RDFAbout}
		err = retrievePredicatesbySubject(conn, recchild.ID, &child, ctx)
		if err != nil {
			return err
		}
		subject.Children = append(subject.Children, child)
	}
	return nil
}