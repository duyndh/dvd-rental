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

//TODO: Need adding tracing to transport layout
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
	r := mux.NewRouter()

	r.Handle("/customer/register", registerHandler)
	return r
}
