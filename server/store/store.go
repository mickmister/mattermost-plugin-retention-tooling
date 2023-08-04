package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
)

type SQLStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
	logger  pluginapi.LogService
}

// New constructs a new instance of SQLStore.
func New(client *pluginapi.Client) (*SQLStore, error) {
	var db *sqlx.DB

	origDB, err := client.Store.GetMasterDB()
	if err != nil {
		return nil, err
	}
	db = sqlx.NewDb(origDB, client.Store.DriverName())

	builder := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	if client.Store.DriverName() == model.DatabaseDriverPostgres {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}

	if client.Store.DriverName() == model.DatabaseDriverMysql {
		db.MapperFunc(func(s string) string { return s })
	}

	return &SQLStore{
		db,
		builder,
		client.Log,
	}, nil
}
