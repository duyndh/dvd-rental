package customer

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
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

func MakeHandler(cs Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	registerHandler := kithttp.NewServer(
		makeRegisterEndpoint(cs),
		decodeRegisterRequest,
		encodeResponse,
		opts...,
	)
	r := mux.NewRouter()
	
	r.Handle("/customer/register", registerHandler)
	return r
}
