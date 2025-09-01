package main

import (
	//"encoding/xml"
	//"github.com/dgraph-io/dgo/v240"
	//"github.com/dgraph-io/dgo/v240/protos/api"
	"cimgraph/dgraph"
	"context"
	"fmt"
	"log"
	"os"
	"github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

type Config struct {
	url string
	configpath string
	path string
}

func main() {
	var config Config

	cli.VersionFlag = &cli.BoolFlag{
		Name: "version",
		Aliases: []string{"v"},
		Usage: "print only the version",
	}

  	cmd := &cli.Command{
		Name: "cimgraph",
		Usage: "Graphing Tool for CIM",
		Version: "development",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "configpath",
				Aliases: []string{"c"},
				Value: "./config/config.yaml",
				Usage: "Location of the cimgraph config file",
				Destination: &config.configpath,
			},
			&cli.StringFlag{
				Name: "url",
				Aliases: []string{"u"},
				Usage: "URL of the Dgraph DB to be used",
				Sources: cli.NewValueSourceChain(yaml.YAML("DgraphURL", altsrc.NewStringPtrSourcer(&config.configpath))),
				Destination: &config.url,
			},
		},
		Commands: []*cli.Command{
			{
				Name: "import",
				Aliases: []string{"i"},
				Usage: "import RDF XML files into the Dgrap DB",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "importpath",
						UsageText: "path of the RDF XML files to import",
						Value: "./data/",
						Destination: &config.path,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
                    if err := importRDF(&config); err != nil {
						return err
					}
                    return nil
                },
			},
			{
				Name: "export",
				Aliases: []string{"e"},
				Usage: "export RDF XML files from the Dgrap DB",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "exportpath",
						UsageText: "path of the RDF XML files to export",
						Value: "./data/",
						Destination: &config.path,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
                    if err := exportRDF(&config); err != nil {
						return err
					}
                    return nil
                },
			},
			{
				Name: "create",
				Usage: "create data schema in the Dgrap DB bases on provided XMI",
				Aliases: []string{"s"},
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "schemapath",
						UsageText: "path of the source XMI XML file",
						Value: "./data/schema.xmi",
						Destination: &config.path,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
                    if err := createSchema(&config); err != nil {
						return err
					}
                    return nil
                },
			},
		},
 	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func importRDF(config *Config) error {
	fmt.Println("importing from: ", config.path , "into Dgraph with URL: ", config.url )
	return nil
}

func exportRDF(config *Config) error {
	fmt.Println("exporting to: ", config.path, "from Dgraph with URL: ", config.url)
	return nil
}

func createSchema(config *Config) error {
	fmt.Println("create schema from", config.path, "into Dgraph with URL: ", config.url)
	err := dgraph.CreateSchema(config.path)
	if err != nil {
		return err
	}
	return nil
}
