package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type RDFS struct {
	ID int64
	RDFType string
	Version string
	Attributes []RDFSAttribute
	Classes []Class
}

type RDFSAttribute struct {
	ID int64
	RDF_ID int64
	Name string
	Value string
}

type Class struct {
	ID int64
	Name string
	Label string
	Comment string
	RDFS_ID int64
	Attributes []Attribute
	Associations []Association
}

type Attribute struct {
	ID int64
	Class_ID int64
	Name string
	Label string
	DataType string
	Mandatory bool
	Comment string
}

type Association struct {
	ID int64
	Parent_ID int64
	Child_ID int64
	Name string
	Label string
	MultiMin int
	MultiMax int
	Comment string
	Range string
	InvereRole string
	Used bool
}

type SubClass struct {
	master_id string
	sub_id string
}

type Enum struct {
	ID int64
	Name string
	Label string
	Comment string
	RDFS_ID int64
}

type EnumValue struct {
	ID int64
	Enum_ID int64
	Name string
	Comment string
}

func InsertRDFS(ptx PostgresTx, rdfs *RDFS, ctx context.Context) error {
	command := "CALL insert_rdfs(@rdftype, @version, @ret_id)"
	args := pgx.NamedArgs{
    		"rdftype": rdfs.RDFType,
    		"version": rdfs.Version,
			"ret_id": nil,
	}
	err := ptx.tx.QueryRow(ctx, command, args).Scan(&rdfs.ID)
	if err != nil {
		return fmt.Errorf("error writing rdfs record: %v", err)
	}
	return nil
}