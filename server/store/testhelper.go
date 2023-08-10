package store

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/store/storetest"
	"github.com/mattermost/mattermost-server/v6/testlib"
)

type TestHelper struct {
	mainHelper *testlib.MainHelper

	Store *SQLStore

	Team1 *model.Team
	Team2 *model.Team
}

func SetupHelper(t *testing.T) *TestHelper {
	var options = testlib.HelperOptions{
		EnableStore: true,
	}

	th := &TestHelper{}
	th.mainHelper = testlib.NewMainHelperWithOptions(&options)

	dbStore := th.mainHelper.GetStore()
	dbStore.DropAllTables()
	dbStore.MarkSystemRanUnitTests()
	th.mainHelper.PreloadMigrations()

	store, err := New(storeWrapper{th.mainHelper}, &testLogger{t})
	require.NoError(t, err, "could not create store")
	th.Store = store

	return th
}

func (th *TestHelper) SetupBasic(t *testing.T) *TestHelper {
	// create some teams
	teams, err := th.createTeams(2, "test-team")
	require.NoError(t, err, "could not create teams")
	th.Team1 = teams[0]
	th.Team2 = teams[1]

	return th
}

func (th *TestHelper) tearDown() {
	if th.mainHelper.SQLStore != nil {
		th.mainHelper.SQLStore.Close()
	}
	if th.mainHelper.Settings != nil {
		storetest.CleanupSqlSettings(th.mainHelper.Settings)
	}
}

func (th *TestHelper) createTeams(num int, namePrefix string) ([]*model.Team, error) {
	var teams []*model.Team
	for i := 0; i < num; i++ {
		team := &model.Team{
			Name:        fmt.Sprintf("%s-%d", namePrefix, i),
			DisplayName: fmt.Sprintf("%s-%d", namePrefix, i),
			Type:        model.TeamOpen,
		}
		team, err := th.mainHelper.Store.Team().Save(team)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, nil
}

// storeWrapper is a wrapper for MainHelper that implemenbts SQLStoreSource interface.
type storeWrapper struct {
	mainHelper *testlib.MainHelper
}

func (sw storeWrapper) GetMasterDB() (*sql.DB, error) {
	return sw.mainHelper.SQLStore.GetInternalMasterDB(), nil
}
func (sw storeWrapper) DriverName() string {
	return *sw.mainHelper.Settings.DriverName
}

type testLogger struct {
	tb testing.TB
}

// Error logs an error message, optionally structured with alternating key, value parameters.
func (l *testLogger) Error(message string, keyValuePairs ...interface{}) {
	l.log("error", message, keyValuePairs...)
}

// Warn logs an error message, optionally structured with alternating key, value parameters.
func (l *testLogger) Warn(message string, keyValuePairs ...interface{}) {
	l.log("warn", message, keyValuePairs...)
}

// Info logs an error message, optionally structured with alternating key, value parameters.
func (l *testLogger) Info(message string, keyValuePairs ...interface{}) {
	l.log("info", message, keyValuePairs...)
}

// Debug logs an error message, optionally structured with alternating key, value parameters.
func (l *testLogger) Debug(message string, keyValuePairs ...interface{}) {
	l.log("debug", message, keyValuePairs...)
}

func (l *testLogger) log(level string, message string, keyValuePairs ...interface{}) {
	var args strings.Builder

	if len(keyValuePairs) > 0 && len(keyValuePairs)%2 != 0 {
		keyValuePairs = keyValuePairs[:len(keyValuePairs)-1]
	}

	for i := 0; i < len(keyValuePairs); i += 2 {
		args.WriteString(fmt.Sprintf("%v:%v  ", keyValuePairs[i], keyValuePairs[i+1]))
	}

	l.tb.Logf("level=%s  message=%s  %s", level, message, args.String())
}
