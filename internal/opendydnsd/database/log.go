package database

import (
	"context"
	"gorm.io/gorm/logger"
	"time"
)

type nopLogger struct {
}

func (np *nopLogger) LogMode(logger.LogLevel) logger.Interface {
	return np
}
func (np *nopLogger) Info(context.Context, string, ...interface{}) {
}
func (np *nopLogger) Warn(context.Context, string, ...interface{}) {
}
func (np *nopLogger) Error(context.Context, string, ...interface{}) {
}
func (np *nopLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
}
