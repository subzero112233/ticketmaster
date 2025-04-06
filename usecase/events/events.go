package events

import (
	"context"
	"github.com/subzero112233/ticketmaster/domain/entity"
)

func NewTicketmasterUseCaseImplementation(repo EventManagementRepository, searchRepo EventManagementSearchRepository, locker LockService) *EventManagementUseCaseImplementation {
	return &EventManagementUseCaseImplementation{
		repository:       repo,
		searchRepository: searchRepo,
		lockService:      locker,
	}
}

func (impl *EventManagementUseCaseImplementation) CreateUser(ctx context.Context, request entity.User) error {
	return impl.repository.CreateUser(ctx, request)
}

func (impl *EventManagementUseCaseImplementation) GetAllEvents(ctx context.Context) ([]entity.Event, error) {
	return impl.repository.GetAllEvents(ctx)
}

func (impl *EventManagementUseCaseImplementation) GetEvent(ctx context.Context, id string) (entity.Event, error) {
	return impl.repository.GetEvent(ctx, id)
}

func (impl *EventManagementUseCaseImplementation) SearchEvents(ctx context.Context, filter *Filter) ([]entity.Event, error) {
	return impl.searchRepository.SearchEvents(ctx, filter)
}

func (impl *EventManagementUseCaseImplementation) GetAvailableTicketsForEvent(ctx context.Context, eventID string) ([]entity.Ticket, error) {
	return impl.repository.GetAvailableTicketsForEvent(ctx, eventID)
}

func (impl *EventManagementUseCaseImplementation) PlaceReservation(ctx context.Context, reservation entity.Reservation) (entity.Reservation, error) {
	// check whether tickets actually exist and belong and belong to the event
	var err error
	reservation.TotalAmount, err = impl.repository.VerifyTicketsAndPrice(ctx, reservation)
	if err != nil {
		return entity.Reservation{}, err
	}

	err = impl.lockService.AcquireLock(ctx, reservation)
	if err != nil {
		return entity.Reservation{}, err
	}

	return impl.repository.PlaceReservation(ctx, reservation)
}
