package customer

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

type loggingService struct {
	logger log.Logger
	Service
}

//NewLoggingSerivce init a logging service
func NewLoggingSerivce(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (l *loggingService) Register(ctx context.Context, name, address string) (err error) {
	defer func(begin time.Time) {
		l.logger.Log("method", "register", "name", name, "address", address, "error", err, "time", time.Since(begin))
	}(time.Now())
	return l.Service.Register(ctx, name, address)
}