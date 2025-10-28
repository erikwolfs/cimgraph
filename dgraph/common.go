package dgraph

import (
	"context"
	"fmt"

	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	URL string
	ImportPath string
	ExportPath string
	DebugPath string
}


func newConnection(config Config) (*dgo.Dgraph, error) {
	conn, err := dgo.NewClient(config.URL,
		
  		// add Dgraph ACL credentials
  		//dgo.WithACLCreds("groot", "password"),
  		// add insecure transport credentials
  	dgo.WithGrpcOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer conn.Close()
	return conn, nil
}

func newTransaction(con *dgo.Dgraph) (*dgo.Txn, error) {
  txn := con.NewTxn()
  return txn, nil
}

func writeMutation(txn *dgo.Txn, rdfdata string) error {
	mu := &api.Mutation{
		SetNquads: []byte(rdfdata),
    	CommitNow: false,
  	}
	if _, err := txn.Mutate(context.Background(), mu); err != nil {
		return err
	}
	return nil
}

func commitTransaction(txn *dgo.Txn) error {
	txn.Commit(context.Background())
	return nil
}

func discardTransaction(txn *dgo.Txn) {
	txn.Discard(context.Background())
}


func dropAllData (conn *dgo.Dgraph) error {
	err := conn.Alter(context.Background(), &api.Operation{DropOp: api.Operation_ALL})
	if err != nil {
		return err
	}
	return nil
}

