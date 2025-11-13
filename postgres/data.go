package postgres

import (
	"context"
	"fmt"
	"time"
	"github.com/jackc/pgx/v5"
)

type DataSet struct {
	ID int64
	TSType string
	RDFAbout string
	Attributes []DSAttribute
}

type DSAttribute struct {
	ID int64
	ATName string
	ATValue string
}

type Subject struct {
	ID int64
	DatasetID int64
	RDFAbout string
	Type string
	ValidFrom time.Time
	ValidTo time.Time
	Predicates []Predicate
}

type ChildSubjectByID struct {
	ID int64 `db:"child_subid"`
	Type string `db:"child_subtype"`
	RDFAbout string `db:"child_rdfid"`
}

type SubjectByRDFID struct {
	ID int64 `db:"subject_id"`
	DatasetID int64 `db:"dataset_id"`
	Type string `db:"subtype"`
	ValidFrom time.Time `db:"valid_from"`
	ValidTo time.Time `db:"valid_to"`
}

type SubjectByType struct {
	ID int64 `db:"subject_id"`
	Type string `db:"subtype"`
	RDFAbout string `db:"rdfid"`
}


type Predicate struct {
	ID int64
	Type string
	Objects []Object
}

type PredicateBySubjectID struct {
	ID int64 `db:"predicate_id"`
	Type string `db:"pretype"`
	ObjectID int64 `db:"object_id"`
	ValueStr string `db:"value_str"`
	ValueFlt float64 `db:"value_flt"`
	ResourceURI string `db:"resource_uri"`
}

type Object struct {
	ID int64
	ValueStr string
	ValueFlt float64
	ResourceURI string
	Resource Resource
}

type EmptyObject struct {
	ID int64 `db:"object_id"`
	ResourceURI string `db:"recourse_uri"`
	ValidFrom time.Time `db:"valid_from"`
	ValidTo time.Time `db:"valid_to"`
}

type ObjectSet struct {
	Objects []Object
}

type Resource struct {
	ObjectID int64
	SubjectID int64
}


func InsertDataSet(ptx PostgresTx, dataset *DataSet, ctx context.Context) error {
	command := "CALL insert_dataset(@tstype, @rdfid, @ret_id)"
	args := pgx.NamedArgs{
    		"tstype": dataset.TSType,
    		"rdfid": dataset.RDFAbout,
			"ret_id": nil,
	}
	err := ptx.tx.QueryRow(ctx, command, args).Scan(&dataset.ID)
	if err != nil {
		return fmt.Errorf("error writing dataset record: %v", err)
	}
	return nil
}

func InsertDSAttribute(ptx PostgresTx, datasetid int64, dsattribute *DSAttribute, ctx context.Context) error {
	command := "CALL insert_dsattribute(@atname, @atvalue, @dataset_id)"
	args := pgx.NamedArgs{
    		"atname": dsattribute.ATName,
    		"atvalue": dsattribute.ATValue,
			"dataset_id": datasetid,
	}
	_, err := ptx.tx.Exec(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error writing dsattribute record: %v", err)
	}
	return nil
}

func InsertSubject(ptx PostgresTx, subject *Subject, ctx context.Context) error {
	command := "CALL insert_subject(@subtype, @dataset_id, @rdfid, @valid_from, @valid_to, @ret_id)"
	args := pgx.NamedArgs{
    		"subtype": subject.Type,
			"dataset_id": subject.DatasetID,
    		"rdfid": subject.RDFAbout,
			"valid_from": "2000-01-01 00:00:00+02",
			"valid_to": "2999-12-31 00:00:00+02",
			"ret_id": nil,
	}
	err := ptx.tx.QueryRow(ctx, command, args).Scan(&subject.ID)
	if err != nil {
		return fmt.Errorf("error writing subject record: %v", err)
	}
	return nil
}

func InsertPredicate(ptx PostgresTx, subjectid int64, predicate *Predicate, ctx context.Context) error {
	command := "CALL insert_predicate(@subject_id, @pretype, @ret_id)"
	args := pgx.NamedArgs{
    		"subject_id": subjectid,
    		"pretype": predicate.Type,
			"ret_id": nil,
	}
	err := ptx.tx.QueryRow(ctx, command, args).Scan(&predicate.ID)
	if err != nil {
		return fmt.Errorf("error writing predicate record: %v", err)
	}
	return nil
}

func InsertObject(ptx PostgresTx, predicateid int64, object *Object, ctx context.Context) error {
	command := "CALL insert_object(@predicate_id, @value_str, @value_flt, @resource_uri, @valid_from, @valid_to, @ret_id)"
	args := pgx.NamedArgs{
    		"predicate_id": predicateid,
    		"value_str": object.ValueStr,
			"value_flt": object.ValueFlt,
			"resource_uri": object.ResourceURI,
			"valid_from": "2000-01-01 00:00:00+02",
			"valid_to": "2999-12-31 00:00:00+02",
			"ret_id": nil,
	}
	err := ptx.tx.QueryRow(ctx, command, args).Scan(&object.ID)
	if err != nil {
		return fmt.Errorf("error writing object record: %v", err)
	}
	return nil
}

func InsertResource(ptx PostgresTx, resource *Resource, ctx context.Context) error {
	command := "CALL insert_resource(@object_id, @subject_id)"
	args := pgx.NamedArgs{
    		"object_id": resource.ObjectID,
    		"subject_id": resource.SubjectID,
	}
	_, err := ptx.tx.Exec(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error writing resource record: %v", err)
	}
	return nil
}

func RetrieveAboutsWithoutResource(ptx PostgresTx, objects *[]*EmptyObject ,ctx context.Context) error {
	command := "select * from retrieve_abouts_without_link()"
	rows, err := ptx.tx.Query(ctx, command)
	if err != nil {
		return fmt.Errorf("error retrieving object abouts without resource: %v", err)
	}
	*objects, err = pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[EmptyObject])
	if err != nil {
		return fmt.Errorf("error parsing object abouts without resource: %v", err)
	}
	return nil
}

func RetrieveSubjectByRDFID(ptx PostgresTx, subjects *[]*SubjectByRDFID, object *EmptyObject, ctx context.Context) error {
	command := "select * from retrieve_subject_by_rdfid(@rdfabout, @validfrom, @validto)"
	args := pgx.NamedArgs{
    		"rdfabout": object.ResourceURI,
    		"validfrom": object.ValidFrom,
			"validto": object.ValidTo,
	}
	rows, err := ptx.tx.Query(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error retrieving subject by rdfid: %v", err)
	}
	*subjects, err = pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SubjectByRDFID])
	if err != nil {
		return fmt.Errorf("error parsing subject after retrieving it by rdfid: %v", err)
	}
	return nil
}

func RetrieveSubjectsByType(con *Postgresdb, subtype *string, subjects *[]*SubjectByType, ctx context.Context) error {
	command := "select * from retrieve_subjects_by_type(@stype, @validfrom, @validto)"
	args := pgx.NamedArgs{
    		"stype": subtype,
    		"validfrom": "2000-01-01 00:00:00+02",
			"validto": "2999-12-31 00:00:00+02",
	}
	rows, err := con.db.Query(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error retrieving subjects by type: %v", err)
	}
	*subjects, err = pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SubjectByType])
	if err != nil {
		return fmt.Errorf("error parsing subjects by type: %v", err)
	}
	return nil
}

func RetrievePredicatesBySubjectID(con *Postgresdb, subid int64, predicates *[]*PredicateBySubjectID, ctx context.Context) error {
	command := "select * from retrieve_predicates_by_subject(@subjectid, @validfrom, @validto)"
	args := pgx.NamedArgs{
    		"subjectid": subid,
    		"validfrom": "2000-01-01 00:00:00+02",
			"validto": "2999-12-31 00:00:00+02",
	}
	rows, err := con.db.Query(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error retrieving predicates by subject id: %v", err)
	}
	*predicates, err = pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PredicateBySubjectID])
	if err != nil {
		return fmt.Errorf("error parsing predicates by subject id: %v", err)
	}
	return nil
}

func RetrieveChildsBySubjectID(con *Postgresdb, subid int64, childsubjects *[]*ChildSubjectByID, ctx context.Context) error {
	command := "select * from retrieve_childs_by_subject(@subjectid, @validfrom, @validto)"
	args := pgx.NamedArgs{
    		"subjectid": subid,
    		"validfrom": "2000-01-01 00:00:00+02",
			"validto": "2999-12-31 00:00:00+02",
	}
	rows, err := con.db.Query(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error retrieving childs by subject id: %v", err)
	}
	*childsubjects, err = pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ChildSubjectByID])
	if err != nil {
		return fmt.Errorf("error parsing childs by subject id: %v", err)
	}
	return nil
}

