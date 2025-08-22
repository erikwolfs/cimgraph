package main

import (
	//"encoding/xml"
	//"github.com/dgraph-io/dgo/v240"
	//"github.com/dgraph-io/dgo/v240/protos/api"
	"context"
	"fmt"
	"log"
	"os"
	"github.com/urfave/cli/v3"
	"github.com/urfave/cli-altsrc/v3"
    yaml "github.com/urfave/cli-altsrc/v3/yaml"
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
		Commands: []*cli.Command{
			{
				Name: "import",
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
                    if err := importrdf(&config); err != nil {
						return err
					}
                    return nil
                },
			},
			{
				Name: "export",
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
                    if err := exportrdf(&config); err != nil {
						return err
					}
                    return nil
                },
			},
		},
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
 	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func importrdf(config *Config) error {
	fmt.Println("importing from: ", config.path)
	return nil
}

func exportrdf(config *Config) error {
	fmt.Println("exporting to: ", config.path)
	return nil
}

