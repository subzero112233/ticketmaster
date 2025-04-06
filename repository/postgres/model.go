package postgres

import "time"

type Event struct {
	Date        time.Time `json:"date" db:"date"`
	ID          string    `json:"id" db:"id"`
	Location    string    `json:"location" db:"location"`
	Name        string    `json:"name" db:"name"`
	Performer   string    `json:"performer" db:"performer"`
	Venue       string    `json:"venue" db:"venue"`
	Description string    `json:"description" db:"description"`
}

type Reservation struct {
	ID          *string  `json:"id,omitempty" db:"id"`
	EventID     string   `json:"event_id" db:"event_id"`
	UserID      string   `json:"user_id" db:"user_id"`
	TicketIDs   []string `json:"tickets" db:"ticket_ids"`
	TotalAmount *float32 `json:"total_amount,omitempty" db:"total_amount"`
}

type Ticket struct {
	ID      string  `json:"id" db:"id"`
	EventID string  `json:"event_id" db:"event_id"`
	Price   float32 `json:"price" db:"price"`
	UserID  *string `json:"user_id" db:"user_id"`
}
