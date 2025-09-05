package dgraph

import (
	"context"
	"fmt"

	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


func newClient() (*dgo.Dgraph, error) {
	client, err := dgo.NewClient("192.168.215.2:9080",
  		// add Dgraph ACL credentials
  		//dgo.WithACLCreds("groot", "password"),
  		// add insecure transport credentials
  	dgo.WithGrpcOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer client.Close()
	return client, nil

}

func dropAllData (conn *dgo.Dgraph) error {
	err := conn.Alter(context.Background(), &api.Operation{DropOp: api.Operation_ALL})
	if err != nil {
		return err
	}
	return nil
}

