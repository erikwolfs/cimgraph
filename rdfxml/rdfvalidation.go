package rdfxml

import (
	db "cimgraph/postgres"
	"context"
	"cimgraph/common"
)

var ValSet map[string]ValProf

type ValProf struct {
	Classes map[string]RDFSClass
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