package database

import (
	"context"
	"github.com/rs/zerolog"
	"gorm.io/gorm/logger"
	"time"
)

type zeroLogger struct {
	logger *zerolog.Logger
}

func (zl *zeroLogger) LogMode(logger.LogLevel) logger.Interface {
	return zl
}
func (zl *zeroLogger) Info(_ context.Context, msg string, data ...interface{}) {
	zl.logger.Trace().Msgf(msg, data...)
}
func (zl *zeroLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	zl.logger.Trace().Msgf(msg, data...)
}
func (zl *zeroLogger) Error(_ context.Context, msg string, data ...interface{}) {
	zl.logger.Error().Msgf(msg, data...)
}
func (zl *zeroLogger) Trace(_ context.Context, _ time.Time, fc func() (string, int64), err error) {
	res, rows := fc()
	zl.logger.Trace().Int64("RowsAffected", rows).Msg(res)
}
