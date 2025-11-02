package main

import (
	"cimgraph/dgraph"
	"cimgraph/postgres"
	"cimgraph/rdfxml"
	"context"
	"fmt"
	"log"
	"os"
	"github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

type config struct {
	dGraphURL string
	configPath string
	path string
	debugPath string
	postgresHost string
	postgresPort string
	postgresDBName string
	postgresUser string
	postgresPassword string
	postgresSchema string
}

func main() {
	var config config

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
				Destination: &config.debugPath,
			},
			&cli.StringFlag{
				Name: "dbhost",
				Aliases: []string{"h"},
				Usage: "Set the postgres database host",
				Sources: cli.NewValueSourceChain(yaml.YAML("PostgresHost", altsrc.NewStringPtrSourcer(&config.configPath))),
				Destination: &config.postgresHost,
			},
			&cli.StringFlag{
				Name: "dbport",
				Aliases: []string{"p"},
				Usage: "Set the postgres database port",
				Sources: cli.NewValueSourceChain(yaml.YAML("PostgresPort", altsrc.NewStringPtrSourcer(&config.configPath))),
				Destination: &config.postgresPort,
			},
			&cli.StringFlag{
				Name: "dbname",
				Aliases: []string{"n"},
				Usage: "Set the postgres database name",
				Sources: cli.NewValueSourceChain(yaml.YAML("PostgresDBName", altsrc.NewStringPtrSourcer(&config.configPath))),
				Destination: &config.postgresDBName,
			},
			&cli.StringFlag{
				Name: "dbuser",
				Aliases: []string{"s"},
				Usage: "Set the postgres database user",
				Sources: cli.NewValueSourceChain(yaml.YAML("PostgresUser", altsrc.NewStringPtrSourcer(&config.configPath))),
				Destination: &config.postgresUser,
			},
			&cli.StringFlag{
				Name: "dbpassword",
				Aliases: []string{"w"},
				Usage: "Set the postgres database user pasword",
				Sources: cli.NewValueSourceChain(yaml.YAML("PostgresPassword", altsrc.NewStringPtrSourcer(&config.configPath))),
				Destination: &config.postgresPassword,
			},
			&cli.StringFlag{
				Name: "dbschema",
				Aliases: []string{"e"},
				Usage: "Set the postgres database schema",
				Sources: cli.NewValueSourceChain(yaml.YAML("PostgresSchema", altsrc.NewStringPtrSourcer(&config.configPath))),
				Destination: &config.postgresSchema,
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
                    if err := importRDFDB(&config); err != nil {
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

func importRDF(config *config) error {
	dconfig := dgraph.Config{URL: config.dGraphURL, ImportPath: config.path, DebugPath: config.debugPath}
	fmt.Println("importing from: ", config.path , "into Dgraph with URL: ", config.dGraphURL )
	err := dgraph.ImportRDF(dconfig)
	if err != nil {
		return err
	}
	return nil
}

func exportRDF(config *config) error {
	fmt.Println("exporting to: ", config.path, "from Dgraph with URL: ", config.dGraphURL)
	return nil
}

func createSchema(config *config) error {
	dconfig := dgraph.Config{URL: config.dGraphURL, ImportPath: config.path, DebugPath: config.debugPath}
	fmt.Println("create schema from", config.path, "into Dgraph with URL: ", config.dGraphURL)
	err := dgraph.CreateSchema(dconfig)
	if err != nil {
		return err
	}
	return nil
}

func importRDFDB(config *config) error {
	dbconfig := postgres.Config{ImportPath: config.path,
								Host: config.postgresHost,
								Port: config.postgresPort,
								DBName: config.postgresDBName,
								User: config.postgresUser,
								Password: config.postgresPassword,
								Schema: config.postgresSchema}
	//err := rdfxml.ImportRDFtoDB(&dbconfig)
	err := rdfxml.Test_empty(&dbconfig)
	if err != nil {
		return err
	}
	return nil
}