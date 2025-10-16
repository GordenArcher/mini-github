package log

import "go.uber.org/zap"

var Logger *zap.Logger

func Init() error {
	l, err := zap.NewProduction()
	if err != nil {
		return err
	}
	Logger = l
	return nil
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
