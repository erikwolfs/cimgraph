package main

import (
	"cimgraph/dgraph"
	"cimgraph/postgres"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

type Config struct {
	dGraphURL string
	configPath string
	path string
	debugpath string
	postgresURL string
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
				Destination: &config.configPath,
			},
			&cli.StringFlag{
				Name: "url",
				Aliases: []string{"u"},
				Usage: "URL of the Dgraph DB to be used",
				Sources: cli.NewValueSourceChain(yaml.YAML("DgraphURL", altsrc.NewStringPtrSourcer(&config.configPath))),
				Destination: &config.dGraphURL,
			},
			&cli.StringFlag{
				Name: "debuglog",
				Aliases: []string{"l"},
				Value: "./log",
				Usage: "Activates writing debug info to the given location",
				Destination: &config.debugpath,
			},
			&cli.StringFlag{
				Name: "dbstring",
				Aliases: []string{"d"},
				Usage: "Set the config string to the postgres database",
				Sources: cli.NewValueSourceChain(yaml.YAML("PostgresConn", altsrc.NewStringPtrSourcer(&config.configPath))),
				Destination: &config.postgresURL,
			},
		},
		Commands: []*cli.Command{
			{
				Name: "importtograph",
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
				Name: "exporttograph",
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
			{
				Name: "importtodb",
				Aliases: []string{"d"},
				Usage: "import RDF XML files into the Postgres DB",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "importpath",
						UsageText: "path of the RDF XML files to import",
						Value: "./data/",
						Destination: &config.path,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
                    if err := importRDFtoDB(&config); err != nil {
						return err
					}
                    return nil
                },
			},
		},
 	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
}

func importRDF(config *Config) error {
	dconfig := dgraph.Config{URL: config.dGraphURL, ImportPath: config.path, DebugPath: config.debugpath}
	fmt.Println("importing from: ", config.path , "into Dgraph with URL: ", config.dGraphURL )
	err := dgraph.ImportRDF(dconfig)
	if err != nil {
		return err
	}
	return nil
}

func exportRDF(config *Config) error {
	fmt.Println("exporting to: ", config.path, "from Dgraph with URL: ", config.dGraphURL)
	return nil
}

func createSchema(config *Config) error {
	dconfig := dgraph.Config{URL: config.dGraphURL, ImportPath: config.path, DebugPath: config.debugpath}
	fmt.Println("create schema from", config.path, "into Dgraph with URL: ", config.dGraphURL)
	err := dgraph.CreateSchema(dconfig)
	if err != nil {
		return err
	}
	return nil
}

func importRDFtoDB(config *Config) error {
	dconfig := postgres.Config{URL: config.postgresURL, ImportPath: config.path, DebugPath: config.debugpath}
	fmt.Println("importing from: ", config.path , "into PostgresSQL with URL: ", config.postgresURL )
	//err := dgraph.ImportRDF(dconfig)
	err := postgres.ImportRDF(dconfig)
	if err != nil {
		return err
	}
	return nil
}
