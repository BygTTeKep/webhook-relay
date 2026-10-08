package logger

import "go.uber.org/zap"

func New(service string, dev bool) (*zap.Logger, error) {
	var (
		l *zap.Logger
		err error
	)
	if dev {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}
	return l.With(zap.String("service", service)), nil
}