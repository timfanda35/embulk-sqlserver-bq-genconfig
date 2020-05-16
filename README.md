# embulk sqlserver to BigQuery configuration template generator

Generate configuration template

```
go run main.go \
  --username=<LOGIN USER> \
  --password=<LOGIN PASSWORD> \
  --hostname=<SQLServer HOST> \
  --database=<DATABASE>
``` 

It reads `INFORMATION_SCHEMA.COLUMNS` table to get schema of tables.

```
.
├── config-20200516_1400
│   ├── PurchaseUsers.yml.liquid
│   └── ThirdpartyVendors.yml.liquid
```

Configuration Template Sample:

```

in:
  type: sqlserver
  host: "{{ env.IN_SOURCE_HOST }}"
  user: "{{ env.IN_SOURCE_DB_USER }}"
  password: "{{ env.IN_SOURCE_DB_PASS }}"
  database: "demo"
  table: "PurchaseUsers"
  select: "create_at, id, name"
  column_options:
    create_at: {type: timestamp}
    id: {type: long}
    name: {type: string}
out:
  type: bigquery
  auth_method: application_default
  project: "{{ env.GCP_PROJECT_ID }}"
  dataset: "{{ env.BQ_DATASET }}"
  table:   "PurchaseUsers"
  compression: GZIP
  gcs_bucket: "{{ env.GCP_PROJECT_ID }}-embulk"
  auto_create_gcs_bucket: true
  column_options:
  - {name: create_at, type: DATETIME, format: '%Y-%m-%d %H:%M:%S'}
  - {name: id, type: INTEGER}
  - {name: name, type: STRING}
```

## Run embulk sample

`task.sh`, remember modify the environment variable.

```
#!/bin/bash

export GCP_PROJECT_ID="<PROJECT_ID>"
export BQ_DATASET="<DATASET>"

export IN_SOURCE_HOST="<SQLServer HOST>"
export IN_SOURCE_DB_USER="<LOGIN USER>"
export IN_SOURCE_DB_PASS="<LOGIN PASSWORD>"
export IN_SOURCE_DB="<DATABASE>"

if [ "$#" -ne 1 ]; then
  echo "Config Dirctory has to be set!"
  exit 1
fi

 echo "################################################"
 echo "Source Host:             $IN_SOURCE_HOST"
 echo "Source DB:               $IN_SOURCE_DB"
 echo "Destination GCP Project: $GCP_PROJECT_ID"
 echo "Destination BQ Dataset:  $BQ_DATASET"
 echo "################################################"

for config in "$1"/*
do
   echo "################################################"
   echo "Processing configuration file: $config"
   echo "################################################"
   embulk run $config
done
```

Execute

```
./task.sh config-20200516_1400
```
