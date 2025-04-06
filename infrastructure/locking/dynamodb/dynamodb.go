package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/subzero112233/ticketmaster/domain/entity"
	"time"
)

type DynamoDB struct {
	client    *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBLocker(db *dynamodb.DynamoDB, tableName string) *DynamoDB {
	return &DynamoDB{client: db, tableName: tableName}
}

func (ddb *DynamoDB) AcquireLock(ctx context.Context, reservation entity.Reservation) error {
	if len(reservation.TicketIDs) == 0 {
		return fmt.Errorf("no ticket IDs provided")
	}

	// DynamoDB supports a max of 25 transactions at a time
	if len(reservation.TicketIDs) > 25 {
		return fmt.Errorf("too many ticket IDs, max allowed is 25")
	}

	ttlValue := time.Now().Add(15 * time.Minute).Unix()
	var transactItems []*dynamodb.TransactWriteItem

	// ensure that the tickets are already locked by the user (in case a previous reservation was interrupted) OR
	// that the tickets are not locked by anyone else, and try to lock them
	for _, ticketID := range reservation.TicketIDs {
		transactItems = append(transactItems, &dynamodb.TransactWriteItem{
			Put: &dynamodb.Put{
				TableName: aws.String(ddb.tableName),
				Item: map[string]*dynamodb.AttributeValue{
					"ticket_id": {S: aws.String(ticketID)},
					"user_id":   {S: aws.String(reservation.UserID)},
					"ttl":       {N: aws.String(fmt.Sprintf("%d", ttlValue))},
				},
				ConditionExpression: aws.String(`
				attribute_not_exists(ticket_id) 
				OR (user_id = :user_id)
			`),
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":user_id": {S: aws.String(reservation.UserID)},
				},
			},
		})
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	}

	_, err := ddb.client.TransactWriteItemsWithContext(ctx, input)
	if err != nil {
		var ddbError *dynamodb.TransactionCanceledException
		if errors.As(err, &ddbError) {
			for _, reason := range ddbError.CancellationReasons {
				if *reason.Code == "ConditionalCheckFailed" {
					return fmt.Errorf("one of the tickets is already being locked for reservation")
				}
			}
		}
		return err
	}

	return nil
}
