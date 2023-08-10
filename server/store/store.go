package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/mattermost/mattermost-server/v6/model"
)

type SQLStoreSource interface {
	GetMasterDB() (*sql.DB, error)
	DriverName() string
}

type Logger interface {
	Error(message string, keyValuePairs ...interface{})
	Warn(message string, keyValuePairs ...interface{})
	Info(message string, keyValuePairs ...interface{})
	Debug(message string, keyValuePairs ...interface{})
}

type SQLStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
	logger  Logger
}

// New constructs a new instance of SQLStore.
func New(src SQLStoreSource, logger Logger) (*SQLStore, error) {
	var db *sqlx.DB

	origDB, err := src.GetMasterDB()
	if err != nil {
		return nil, err
	}
	db = sqlx.NewDb(origDB, src.DriverName())

	builder := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	if src.DriverName() == model.DatabaseDriverPostgres {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}

	if src.DriverName() == model.DatabaseDriverMysql {
		db.MapperFunc(func(s string) string { return s })
	}

	return &SQLStore{
		db,
		builder,
		logger,
	}, nil
}
