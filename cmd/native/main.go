package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/jmoiron/sqlx"
	"github.com/subzero112233/ticketmaster/api/chi/handler"
	dynamodbLocker "github.com/subzero112233/ticketmaster/infrastructure/locking/dynamodb"
	elasticsearchRepository "github.com/subzero112233/ticketmaster/repository/elasticsearch"
	"github.com/subzero112233/ticketmaster/repository/postgres"
	"github.com/subzero112233/ticketmaster/repository/postgres/migrations"
	"github.com/subzero112233/ticketmaster/usecase/events"
	"net/http"
	"time"
)

const (
	dynamodbTable      = "ticket_locks"
	region             = "us-east-1"
	localstackEndpoint = "http://localstack:4566"
)

func main() {
	ctx := context.Background()

	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s port=%d", "postgres", "myuser", "mypassword", "mydatabase", "disable", 5432) // uses docker-compose network
	databaseClient, err := sqlx.ConnectContext(ctx, "postgres", connectionString)
	if err != nil {
		panic(err)
	}

	// run migrations. on production this should be done as part of the deployment process
	_, err = databaseClient.ExecContext(ctx, migrations.Migrations)
	if err != nil {
		panic(err)
	}

	// initialize the repository implementation
	repository := postgres.NewPostgresRepository(databaseClient)

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://elasticsearch:9200", // uses docker-compose network
		},
		Transport: &http.Transport{
			Proxy:              http.ProxyFromEnvironment,
			DisableCompression: true, // disable compression globally
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	// initialize a search repository
	elasticsearchRepo := elasticsearchRepository.NewElasticSearchImplementation(client, "events-table-topic.public.events")

	// initialize the distributed lock service
	sess, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String(localstackEndpoint),
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	})

	if err != nil {
		panic(err)
	}

	lockerService := dynamodbLocker.NewDynamoDBLocker(dynamodb.New(sess), dynamodbTable)

	// initialize the use case implementation
	usecaseImplementation := events.NewTicketmasterUseCaseImplementation(repository, elasticsearchRepo, lockerService)

	serverHandler, err := handler.NewChiHandler(usecaseImplementation)
	if err != nil {
		panic("could not initialize handler with error: " + err.Error())
	}

	server := &http.Server{
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
		Addr:              ":8000",
		ReadHeaderTimeout: 30 * time.Second,
		Handler:           serverHandler,
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
