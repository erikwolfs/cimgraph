package como

import (
	db "cimgraph/postgres"
	"context"
	"encoding/json"
	"fmt"
	"cimgraph/common"
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
	t := common.CurrentTime()
	ctx := context.Background()
	pgpool, err := db.NewConnectionPool(config, ctx)
 	if err != nil {
 		return err
 	}
	pgcon, err := pgpool.NewConnection(ctx)
	if err != nil {
		return err
	}
	err = pgcon.SetSchema(config.Schema, ctx)
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
	pgcon.Release()
	pgpool.Close()
	fmt.Print(string(b))
	common.MeasureTime("write JSON output", t)
	return nil
}

func retrievePredicatesbySubject(conn *db.PostgresConn, subid int64, subject *Subject, ctx context.Context) error {
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

func retrieveChildsbySubject(conn *db.PostgresConn, subid int64, subject *Subject, ctx context.Context) error {
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