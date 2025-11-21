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
	Items []RDFSItem
	Classes []Class
	Enums []Enum
}

type RDFSItem struct {
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

func InsertRDFSItem(ptx PostgresTx, item *RDFSItem, ctx context.Context) error {
	command := "CALL insert_rdfsattribute(@atname, @atvalue, @rdfs_id)"
	args := pgx.NamedArgs{
    		"atname": item.Name,
    		"atvalue": item.Value,
			"rdfs_id": item.RDF_ID,
	}
	_, err := ptx.tx.Exec(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error writing rdfs item record: %v", err)
	}
	return nil
}

func InsertEnum(ptx PostgresTx, enum *Enum, ctx context.Context) error {
	command := "CALL insert_enum(@enname, @enlabel, @encomment, @rdfsid, @ret_id)"
	args := pgx.NamedArgs{
    		"enname": enum.Name,
    		"enlabel": enum.Label,
			"encomment": enum.Comment,
			"rdfsid": enum.RDFS_ID,
			"ret_id": nil,
	}
	err := ptx.tx.QueryRow(ctx, command, args).Scan(&enum.ID)
	if err != nil {
		return fmt.Errorf("error writing enum record: %v", err)
	}
	return nil
}

func InsertEnumValue(ptx PostgresTx, value *EnumValue, ctx context.Context) error {
	command := "CALL insert_enumvalue(@enumid, @evname, @evcomment)"
	args := pgx.NamedArgs{
    		"enumid": value.Enum_ID,
    		"evname": value.Name,
			"evcomment": value.Comment,
	}
	_, err := ptx.tx.Exec(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error writing enum value record: %v", err)
	}
	return nil
}

func InsertClass(ptx PostgresTx, class *Class, ctx context.Context) error {
	command := "CALL insert_class(@clname, @cllabel, @clcomment, @rdfsid, @ret_id)"
	args := pgx.NamedArgs{
    		"clname": class.Name,
    		"cllabel": class.Label,
			"clcomment": class.Comment,
			"rdfsid": class.RDFS_ID,
			"ret_id": nil,
	}
	err := ptx.tx.QueryRow(ctx, command, args).Scan(&class.ID)
	if err != nil {
		return fmt.Errorf("error writing class record: %v", err)
	}
	return nil
}

func InsertAssociation(ptx PostgresTx, asso *Association, ctx context.Context) error {
	command := "CALL insert_association(@parentid, @childid, @asname, @aslabel, @munin, @mumax, @ascomment, @asrange, @inverserole, @used)"
	args := pgx.NamedArgs{
    		"parentid": asso.Parent_ID,
    		"childid": asso.Child_ID,
			"asname": asso.Name,
			"aslabel": asso.Label,
			"mumin": asso.MultiMin,
			"mumax": asso.MultiMax,
			"ascomment": asso.Comment,
			"asrange": asso.Range,
			"inverserole": asso.InvereRole,
			"used": asso.Used,
	}
	_, err := ptx.tx.Exec(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error writing association record: %v", err)
	}
	return nil
}

func InsertAttribute(ptx PostgresTx, attr *Attribute, ctx context.Context) error {
	command := "CALL insert_attribute(@classid, @atname, @atlabel, @datatype, @atcomment, @mandatory)"
	args := pgx.NamedArgs{
    		"classid": attr.Class_ID,
			"atname": attr.Name,
			"atlabel": attr.Label,
			"datatype": attr.DataType,
			"atcomment": attr.Comment,
			"mandatory": attr.Mandatory,
	}
	_, err := ptx.tx.Exec(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error writing attribute record: %v", err)
	}
	return nil
}

func InsertSubClass(ptx PostgresTx, subcl *SubClass, ctx context.Context) error {
	command := "CALL insert_subclass(@masterid, @subid)"
	args := pgx.NamedArgs{
    		"masterid": subcl.master_id,
    		"subid": subcl.sub_id,
	}
	_, err := ptx.tx.Exec(ctx, command, args)
	if err != nil {
		return fmt.Errorf("error writing subclass record: %v", err)
	}
	return nil
}