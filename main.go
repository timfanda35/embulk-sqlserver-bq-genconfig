package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	mssql "github.com/denisenkom/go-mssqldb"
	"log"
	"net/url"
	"os"
	"text/template"
	"time"
)

const SQL_QUERY = "SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS ORDER BY TABLE_NAME, COLUMN_NAME"
const OUTPUT_TEMPLATE = `
in:
  type: sqlserver
  host: {{"{{ env.IN_SOURCE_HOST }}"}}
  user: {{"{{ env.IN_SOURCE_DB_USER }}"}}
  password: {{"{{ env.IN_SOURCE_DB_PASS }}"}}
  database: {{"{{ env.IN_SOURCE_DB }}"}}
  table: {{"{{ env.IN_SOURCE_TABLE }}"}}
  select: "*"
  column_options:
  {{- range . }}
  - {{ .InputDefine }}
  {{- end }}
out:
  type: bigquery
  auth_method: application_default
  project: {{"{{ env.GCP_PROJECT_ID }}"}}
  dataset: {{"{{ env.BQ_DATASET }}"}}
  table:   {{"{{ env.BQ_TABLE }}"}}
  compression: GZIP
  gcs_bucket: {{"{{ env.GCP_PROJECT_ID }}-embulk"}}
  auto_create_gcs_bucket: true
  column_options:
  {{- range . }}
  - {{ .OuputDefine }}
  {{- end }}
`

type TableSchemaColumn struct {
	Name     string
	DataType string
}

func (t TableSchemaColumn) InputDefine() string {
	var embulkType string

	switch t.DataType {
	case "int", "smallint", "bigint":
		embulkType = "long"
	case "date", "time", "datetime", "timestamp":
		embulkType = "timestamp"
	default:
		embulkType = "string"
	}

	return fmt.Sprintf("%s: {type: %s}", t.Name, embulkType)
}

func (t TableSchemaColumn) OuputDefine() string {
	var bqType string

	switch t.DataType {
	case "int", "smallint", "bigint":
		bqType = "INTEGER"
	case "date":
		bqType = "DATETIME, format: '%Y-%m-%d'"
	case "time":
		bqType = "DATETIME, format: '%H:%M:%S'"
	case "datetime":
		bqType = "DATETIME, format: '%Y-%m-%d %H:%M:%S'"
	case "timestamp":
		bqType = "timestamp"
	default:
		bqType = "string"
	}

	return fmt.Sprintf("{name: %s, type: %s}", t.Name, bqType)
}

func getConnector(username, password, hostname, database string, port int) *mssql.Connector {
	query := url.Values{}
	query.Add("database", database)

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(username, password),
		Host:     fmt.Sprintf("%s:%d", hostname, port),
		RawQuery: query.Encode(),
	}

	connector, err := mssql.NewConnector(u.String())
	if err != nil {
		log.Fatal(err)
	}

	return connector
}

func parseRows(rows *sql.Rows) map[string][]TableSchemaColumn {
	tableSchemas := make(map[string][]TableSchemaColumn)

	var (
		tableName  string
		columnName string
		columnType string
	)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&tableName, &columnName, &columnType)
		if err != nil {
			log.Fatal(err)
		}

		if len(tableSchemas[tableName]) == 0 {
			tableSchemas[tableName] = []TableSchemaColumn{}
		}
		tableSchemas[tableName] = append(
			tableSchemas[tableName],
			TableSchemaColumn{
				Name:     columnName,
				DataType: columnType,
			})
	}

	return tableSchemas
}

func output(tableSchemas map[string][]TableSchemaColumn) {
	// output
	outputDir := fmt.Sprintf("config-%s", time.Now().Format("20060102_1504"))
	err := os.Mkdir(outputDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.New("template").Parse(OUTPUT_TEMPLATE)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range tableSchemas {
		outputPath := fmt.Sprintf("%s/%s.yml.liquid", outputDir, k)
		f, err := os.Create(outputPath)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf(">>>> %s <<<<", k)
		if err := tmpl.Execute(f, v); err != nil {
			f.Close()
			log.Fatal(err)
		}

		f.Close()
		log.Printf("-------- wrote to %s", outputPath)
	}
}

func process(username, password, hostname, database string, port int) {
	connector := getConnector(username, password, hostname, database, port)
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rows, err := db.QueryContext(ctx, SQL_QUERY)
	if err != nil {
		log.Fatal(err)
	}
	tableSchemas := parseRows(rows)

	output(tableSchemas)

	log.Printf("==> Done")
}

func main() {
	username := flag.String("username", "sa", "username")
	password := flag.String("password", "", "password")
	hostname := flag.String("hostname", "127.0.0.1", "hostname")
	database := flag.String("database", "", "database")
	port := flag.Int("port", 1433, "port")

	flag.Parse()

	if *password == "" {
		log.Fatal("Must set password")
	}

	if *database == "" {
		log.Fatal("Must set database")
	}

	process(*username, *password, *hostname, *database, *port)
}
