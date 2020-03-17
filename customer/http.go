package customer

import (
	"context"
	"encoding/json"
	"net/http"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	stdopentracing "github.com/opentracing/opentracing-go"
)

func decodeRegisterRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return registerRequest{
		Name:    body.Name,
		Address: body.Address,
	}, nil
}

func decodeRentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		CustomerID string `json:"customer_id"`
		DVDID      string `json:"dvd_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return rentRequest{
		CustomerID:    body.CustomerID,
		DVDID: body.DVDID,
	}, nil
}

type errorer interface {
	error() error
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	switch err {
	case errInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func MakeHandler(endpoints CustomerEndpoints, logger kitlog.Logger, ot stdopentracing.Tracer) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	registerHandler := kithttp.NewServer(
		endpoints.RegisterEndpoint,
		decodeRegisterRequest,
		encodeResponse,
		append(opts, kithttp.ServerBefore(opentracing.HTTPToContext(ot, "register", logger)))...,
	)

	rentHandler := kithttp.NewServer(
		endpoints.RentEndpoint,
		decodeRentRequest,
		encodeResponse,
		append(opts, kithttp.ServerBefore(opentracing.HTTPToContext(ot, "rent", logger)))...,
	)

	r := mux.NewRouter()

	r.Handle("/customer/v1/register", registerHandler)
	r.Handle("/customer/v1/rent", rentHandler)
	return r
}
