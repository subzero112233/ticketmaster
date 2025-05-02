#!/bin/bash

set -e

export AWS_PAGER=""

echo "starting docker-compose"
docker-compose up --build -d

echo "waiting for services to start"
sleep 10

echo "creating dynamodb table"
aws --endpoint-url=http://localhost:4566 cloudformation create-stack --stack-name ticketmaster --template-body file://deploy/dynamodb.yaml --capabilities CAPABILITY_AUTO_EXPAND CAPABILITY_NAMED_IAM

echo "creating kafka connectors (make sure kafka-connect already started)"
curl -XPOST --location 'http://localhost:8083/connectors' \
--header 'Accept: application/json' \
--header 'Content-Type: application/json' \
--data '{
"name": "cdc-debezium-connector-postgres",
"config": {
"connector.class": "io.debezium.connector.postgresql.PostgresConnector",
"database.hostname": "host.docker.internal",
"database.port": "5432",
"database.user": "myuser",
"database.password": "mypassword",
"database.dbname": "mydatabase",
"database.server.id": "122054",
"table.include.list": "public.events",
"topic.prefix": "events-table-topic",
"transforms": "unwrap",
"transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
"transforms.unwrap.drop.tombstones": "true",
"transforms.unwrap.delete.handling.mode": "drop",
"value.converter": "org.apache.kafka.connect.json.JsonConverter",
"value.converter.schemas.enable": "false",
"key.converter": "org.apache.kafka.connect.json.JsonConverter",
"key.converter.schemas.enable": "false",
"database.history.kafka.topic": "dbhistory.mydb",
"database.history.kafka.bootstrap.servers": "kafka:9092"
}
}
'

echo "Creating fixed elasticsearch sink connector (using id as _id)..."

curl -X POST http://localhost:8083/connectors \
--header 'Content-Type: application/json' \
--data '{
  "name": "elasticsearch-sink-connector",
  "config": {
    "connector.class": "io.confluent.connect.elasticsearch.ElasticsearchSinkConnector",
    "tasks.max": "1",
    "topics": "events-table-topic.public.events",
    "connection.url": "http://elasticsearch:9200",
    "type.name": "_doc",
    "key.ignore": "false",
    "schema.ignore": "true",
    "insert.mode": "upsert",
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter.schemas.enable": "false",
    "transforms": "unwrap,extractKey",
    "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
    "transforms.unwrap.drop.tombstones": "true",
    "transforms.unwrap.delete.handling.mode": "drop",
    "transforms.extractKey.type": "org.apache.kafka.connect.transforms.ExtractField$Key",
    "transforms.extractKey.field": "id"
  }
}'
