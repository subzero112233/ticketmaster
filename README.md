# Ticketmaster-like application
In this repository, I demonstrate a basic cloud version of the famous Ticketmaster website with the primary
goal of focusing on 2 advanced techniques:
1. Distributed Lock - Providing a decent user experience when booking, by using a distributed lock, implemented using DynamoDB with LocalStack
   Meaning you can run the application locally and test it without needing to deploy it to AWS.
2. Change Data Capture (CDC) - Enabling advanced search capabilities through a robust CDC procedure, implemented using Debezium and Kafka.
   This allows you to stream real-time incremental changes from an OLTP database (Postgres) to a Search Engine (ElasticSearch) and allow advanced search capabilities for events based on various criteria, such as date, location, and performer and most importantly, description

I used my own skeleton CLI https://github.com/skyhawk-security/goskeleton (with slight modifications) to generate a skeleton service with Clean Architecture, AWS Serverless deployment (which was discarded for this project) and more.

## Prerequisites
1. Docker - https://docs.docker.com/engine/install/
2. docker-compose - https://docs.docker.com/compose/install/

## Installation
```bash
./start.sh
```

## Get all events
curl --location 'http://localhost:8000/events'

## Get a specific event
curl --location 'http://localhost:8000/events/$id'

## Search for an event
curl --location 'http://localhost:8000/events/search?description=metal%20legends&from_date=1900-04-06T00%3A00%3A00Z&Performer=Pantera&location=New%20York'

## Get available tickets for event
curl --location 'http://localhost:8000/events/$id/tickets'

## Reserve a ticket
curl --location 'http://localhost:8000/reservations' \
--header 'user-email: reshefsharvit21@gmail.com' \
--header 'Content-Type: application/json' \
--data '{"event_id": ["<EVENT_ID_HERE>"], "tickets": ["<TICKET_ID_HERE>"]}'
