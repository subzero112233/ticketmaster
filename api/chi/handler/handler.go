package handler

import (
	"context"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/subzero112233/ticketmaster/api"
	"github.com/subzero112233/ticketmaster/usecase/events"
	"net/http"
)

type ChiHandler struct {
	UseCase events.EventManagement
}

func (c ChiHandler) CreateUser(ctx context.Context, request api.CreateUserRequestObject) (api.CreateUserResponseObject, error) {
	err := c.UseCase.CreateUser(ctx, toEntityUser(*request.Body))
	if err != nil {
		return api.CreateUserdefaultJSONResponse{
			Body:       api.ErrorOutput{Message: err.Error()},
			StatusCode: 500,
		}, nil
	}

	return api.CreateUser201Response{}, nil
}

// no auth needed
func (c ChiHandler) GetAvailableTicketsForEvent(ctx context.Context, request api.GetAvailableTicketsForEventRequestObject) (api.GetAvailableTicketsForEventResponseObject, error) {
	availableTickets, err := c.UseCase.GetAvailableTicketsForEvent(ctx, request.EventId)
	if err != nil {
		return api.GetAvailableTicketsForEventdefaultJSONResponse{
			Body:       api.ErrorOutput{Message: err.Error()},
			StatusCode: 500,
		}, nil
	}

	return api.GetAvailableTicketsForEvent200JSONResponse(toTicketsResponse(availableTickets)), nil
}

// no auth needed
func (c ChiHandler) GetAllEvents(ctx context.Context, _ api.GetAllEventsRequestObject) (api.GetAllEventsResponseObject, error) {
	allEvents, err := c.UseCase.GetAllEvents(ctx)
	if err != nil {
		return api.GetAllEventsdefaultJSONResponse{
			Body:       api.ErrorOutput{Message: err.Error()},
			StatusCode: 500,
		}, nil
	}

	return api.GetAllEvents200JSONResponse(toEventsResponse(allEvents)), nil
}

// no auth needed
func (c ChiHandler) GetEvent(ctx context.Context, request api.GetEventRequestObject) (api.GetEventResponseObject, error) {
	event, err := c.UseCase.GetEvent(ctx, request.EventId)
	if err != nil {
		return api.GetEventdefaultJSONResponse{
			Body:       api.ErrorOutput{Message: err.Error()},
			StatusCode: 500,
		}, nil
	}

	return api.GetEvent200JSONResponse(toEventResponse(event)), nil
}

func (c ChiHandler) SearchEvents(ctx context.Context, request api.SearchEventsRequestObject) (api.SearchEventsResponseObject, error) {
	events, err := c.UseCase.SearchEvents(ctx, toFilter(request))
	if err != nil {
		return api.SearchEventsdefaultJSONResponse{
			Body:       api.ErrorOutput{Message: err.Error()},
			StatusCode: 500,
		}, nil
	}

	return api.SearchEvents200JSONResponse(toEventsResponse(events)), nil
}

func (c ChiHandler) PlaceReservation(ctx context.Context, request api.PlaceReservationRequestObject) (api.PlaceReservationResponseObject, error) {
	userEmail, ok := ctx.Value("user-email").(string)
	if !ok {
		return api.PlaceReservationdefaultJSONResponse{
			Body:       api.ErrorOutput{Message: "user id not present in request"},
			StatusCode: 400,
		}, nil
	}

	reservation, err := c.UseCase.PlaceReservation(ctx, toReservation(request, userEmail))
	if err != nil {
		return api.PlaceReservationdefaultJSONResponse{
			Body:       api.ErrorOutput{Message: err.Error()},
			StatusCode: 500,
		}, nil
	}

	return toReservationResponse(reservation), nil
}

func NewChiHandler(usecase events.EventManagement) (http.Handler, error) {
	handlerFunctions := &ChiHandler{
		UseCase: usecase,
	}

	swaggerSpec, err := api.GetSwagger()
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(nethttpmiddleware.OapiRequestValidator(swaggerSpec))

	r.Use(AddCustomHeadersAndContextObjects)
	r.Use(VerifyMandatoryContextObjects())

	// NOTE: Hover over handlerFunctions and implement the missing methods
	strictHandler := api.NewStrictHandler(handlerFunctions, nil)

	// IMPORTANT: make sure that you have successfully generated the server through the OpenAPI generator. otherwise, it won't work.
	return api.HandlerFromMux(strictHandler, r), nil
}
