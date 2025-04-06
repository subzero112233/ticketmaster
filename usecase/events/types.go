package events

import (
	"context"
	"github.com/subzero112233/ticketmaster/domain/entity"
	"time"
)

// UseCase could be in a separate directory, as implemented here. but in case it's the only use case. it could also be placed directly in the usecase dir
type EventManagement interface {
	CreateUser(ctx context.Context, request entity.User) error
	GetAllEvents(ctx context.Context) ([]entity.Event, error)
	GetEvent(ctx context.Context, id string) (entity.Event, error)
	SearchEvents(ctx context.Context, filter *Filter) ([]entity.Event, error)
	GetAvailableTicketsForEvent(ctx context.Context, eventID string) ([]entity.Ticket, error)
	PlaceReservation(ctx context.Context, reservation entity.Reservation) (entity.Reservation, error)
}

// this is the use case implementation.
// notice the lowercase repository. it means it's unexported and therefor cannot be accessed.
type EventManagementUseCaseImplementation struct {
	repository       EventManagementRepository
	searchRepository EventManagementSearchRepository
	lockService      LockService
}

type EventManagementRepository interface {
	CreateUser(ctx context.Context, request entity.User) error
	GetAllEvents(ctx context.Context) ([]entity.Event, error)
	GetEvent(ctx context.Context, id string) (entity.Event, error)
	GetAvailableTicketsForEvent(ctx context.Context, eventID string) ([]entity.Ticket, error)
	PlaceReservation(ctx context.Context, reservation entity.Reservation) (entity.Reservation, error)
	VerifyTicketsAndPrice(ctx context.Context, reservation entity.Reservation) (float32, error)
}

type EventManagementSearchRepository interface {
	SearchEvents(ctx context.Context, filter *Filter) ([]entity.Event, error)
}

type LockService interface {
	AcquireLock(ctx context.Context, reservation entity.Reservation) error
}

type Filter struct {
	Performer   *string
	Location    *string
	Description *string
	FromDate    time.Time
	ToDate      time.Time
	Page        int
}
