package main

import (
	"cimgraph/como"
	"cimgraph/postgres"
	"cimgraph/rdfxml"
	"cimgraph/common"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

type config struct {
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
				Name: "exporttograph",
				Aliases: []string{"e"},
				Usage: "export JSON graph message from the DB",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "subject type",
						UsageText: "give the root subject type for the graph",
						Value: "PowerTransformer/",
						Destination: &config.path,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
                    if err := exportRDFGraph(&config); err != nil {
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
			{
				Name: "importRDFschematodb",
				Aliases: []string{"s"},
				Usage: "import RDFS XML schema files into the Postgres DB",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "importpath",
						UsageText: "path of the RDFS XML files to import",
						Value: "./data/",
						Destination: &config.path,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
                    if err := importRDFSDB(&config); err != nil {
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

func exportRDFGraph(config *config) error {
	fmt.Println("exporting Graph for: ", config.path)
	dbconfig := postgres.Config{ImportPath: config.path,
								Host: config.postgresHost,
								Port: config.postgresPort,
								DBName: config.postgresDBName,
								User: config.postgresUser,
								Password: config.postgresPassword,
								Schema: config.postgresSchema}
	err := como.GenerateGraphBySubjectType(&dbconfig, config.path)
	if err != nil {
		return err
	}
	return nil
}


func importRDFDB(config *config) error {
	defer exeTime("importRDFDB")
	dbconfig := postgres.Config{ImportPath: config.path,
								Host: config.postgresHost,
								Port: config.postgresPort,
								DBName: config.postgresDBName,
								User: config.postgresUser,
								Password: config.postgresPassword,
								Schema: config.postgresSchema}
	err := rdfxml.ImportRDFtoDB(&dbconfig)
	if err != nil {
		return err
	}
	return nil
}

func importRDFSDB(config *config) error {
	t := common.CurrentTime()
	defer common.MeasureTime("importDFSDB", t)
	dbconfig := postgres.Config{ImportPath: config.path,
								Host: config.postgresHost,
								Port: config.postgresPort,
								DBName: config.postgresDBName,
								User: config.postgresUser,
								Password: config.postgresPassword,
								Schema: config.postgresSchema}
	err := rdfxml.ImportRDFStoDB(&dbconfig)
	if err != nil {
		return err
	}
	return nil
}

func exeTime(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s execution time: %v\n", name, time.Since(start))
	}
}