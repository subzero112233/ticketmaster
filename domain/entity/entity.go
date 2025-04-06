package entity

import "time"

type Event struct {
	Date        time.Time `json:"date"`
	ID          string    `json:"id"`
	Location    string    `json:"location"`
	Name        string    `json:"name"`
	Performer   string    `json:"performer"`
	Venue       string    `json:"venue"`
	Description string    `json:"description"`
}

type Venue struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Capacity int    `json:"capacity"`
}

type Performer struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type User struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Ticket struct {
	ID      string  `json:"id"`
	EventID string  `json:"event_id"`
	Price   float32 `json:"price"`
	UserID  *string `json:"user_id"`
}

type Reservation struct {
	ID          *string    `json:"id,omitempty"`
	Date        *time.Time `json:"date,omitempty"`
	EventID     string     `json:"event_id"`
	UserID      string     `json:"user_id"` // references to email
	TicketIDs   []string   `json:"tickets"`
	TotalAmount float32    `json:"total_amount,omitempty"`
}
