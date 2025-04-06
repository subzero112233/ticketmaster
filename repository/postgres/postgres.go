package postgres

import (
	"context"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/subzero112233/ticketmaster/domain/entity"
	"log/slog"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (pgdb *PostgresRepository) CreateUser(ctx context.Context, user entity.User) error {
	query, args, err := goqu.Insert("users").Rows(goqu.Record{
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}).ToSQL()
	if err != nil {
		slog.Error(fmt.Sprintf("unable to generate sql query with error: %s", err.Error()))
		return err
	}

	_, err = pgdb.db.ExecContext(ctx, query, args...)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to create user with error: %s", err.Error()))
		return err
	}

	return nil
}

func (pgdb *PostgresRepository) GetAllEvents(ctx context.Context) ([]entity.Event, error) {
	var eventResult []Event
	err := pgdb.db.SelectContext(ctx, &eventResult, "SELECT * FROM events")
	if err != nil {
		slog.Error(fmt.Sprintf("unable to get eventResult with error: %s", err.Error()))
		return nil, err
	}

	var result []entity.Event
	for _, event := range eventResult {
		result = append(result, entity.Event(event))
	}

	return result, nil
}

func (pgdb *PostgresRepository) GetEvent(ctx context.Context, id string) (entity.Event, error) {
	query, args, err := goqu.From("events").Where(goqu.Ex{"id": id}).ToSQL()
	if err != nil {
		slog.Error(fmt.Sprintf("unable to generate sql query with error: %s", err.Error()))
		return entity.Event{}, err
	}

	var event Event
	err = pgdb.db.GetContext(ctx, &event, query, args...)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to get event with error: %s", err.Error()))
		return entity.Event{}, err
	}

	return entity.Event(event), nil
}

func (pgdb *PostgresRepository) GetAvailableTicketsForEvent(ctx context.Context, eventID string) ([]entity.Ticket, error) {
	query, args, err := goqu.From("tickets").Where(goqu.Ex{"event_id": eventID, "user_id": nil}).ToSQL()
	if err != nil {
		slog.Error(fmt.Sprintf("unable to generate sql query with error: %s", err.Error()))
		return nil, err
	}

	var ticketResult []Ticket
	err = pgdb.db.SelectContext(ctx, &ticketResult, query, args...)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to get ticketResult with error: %s", err.Error()))
		return nil, err
	}

	var result []entity.Ticket
	for _, ticket := range ticketResult {
		result = append(result, entity.Ticket(ticket))
	}

	return result, nil
}

func (pgdb *PostgresRepository) PlaceReservation(ctx context.Context, reservation entity.Reservation) (entity.Reservation, error) {
	// start a transaction
	tx, err := pgdb.db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to start transaction with error: %s", err.Error()))
		return entity.Reservation{}, err
	}

	// ensure all tickets are unassigned
	var count int
	err = tx.GetContext(ctx, &count, "SELECT COUNT(*) FROM tickets WHERE id = ANY($1) AND user_id IS NULL", pq.Array(reservation.TicketIDs))
	if err != nil {
		slog.Error(fmt.Sprintf("unable to check ticket status with error: %s", err.Error()))
		return entity.Reservation{}, err
	}

	if count != len(reservation.TicketIDs) {
		slog.Error(fmt.Sprintf("some tickets are already assigned"))
		return entity.Reservation{}, fmt.Errorf("some tickets are already assigned")
	}

	// assign tickets to user
	query, args, err := goqu.Update("tickets").Set(goqu.Record{"user_id": reservation.UserID}).Where(goqu.Ex{"event_id": reservation.EventID, "user_id": nil}).Where(goqu.C("id").In(reservation.TicketIDs)).ToSQL()
	if err != nil {
		slog.Error(fmt.Sprintf("unable to generate sql query with error: %s", err.Error()))
		return entity.Reservation{}, err
	}

	_, err = pgdb.db.ExecContext(ctx, query, args...)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to assign tickets with error: %s", err.Error()))
		return entity.Reservation{}, err
	}

	// place reservation
	dialect := goqu.Dialect("postgres")
	query, args, err = dialect.Insert("reservations").
		Rows(goqu.Record{
			"event_id":     reservation.EventID,
			"user_id":      reservation.UserID,
			"ticket_ids":   pq.Array(reservation.TicketIDs), // Important for UUID[]
			"total_amount": reservation.TotalAmount,
		}).Returning("date").
		ToSQL()

	if err != nil {
		slog.Error(fmt.Sprintf("unable to generate sql query with error: %s", err.Error()))
		return entity.Reservation{}, err
	}

	err = pgdb.db.QueryRowContext(ctx, query, args...).Scan(&reservation.Date)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to place reservation with error: %s", err.Error()))
		return entity.Reservation{}, err
	}

	return reservation, nil
}

func (pgdb *PostgresRepository) VerifyTicketsAndPrice(ctx context.Context, reservation entity.Reservation) (float32, error) {
	var tickets []Ticket
	err := pgdb.db.SelectContext(ctx, &tickets, "SELECT * FROM tickets WHERE id = ANY($1) AND event_id = $2 AND user_id IS NULL", pq.Array(reservation.TicketIDs), reservation.EventID)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to check ticket existence with error: %s", err.Error()))
		return 0, err
	}

	if len(tickets) != len(reservation.TicketIDs) {
		return 0, fmt.Errorf("some tickets do not exist or are already assigned")
	}

	var totalAmount float32
	for _, ticket := range tickets {
		totalAmount += ticket.Price
	}

	return totalAmount, nil
}
