package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-retention-tooling/server/command"
	"github.com/mattermost/mattermost-plugin-retention-tooling/server/config"
	"github.com/mattermost/mattermost-plugin-retention-tooling/server/jobs"
	"github.com/mattermost/mattermost-plugin-retention-tooling/server/store"
)

const (
	routeRemoveUserFromAllTeamsAndChannels = "/remove_user_from_all_teams_and_channels"
	ChannelArchiverJobID                   = "channel_archiver_job"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	configuration     *config.Configuration

	Client   *pluginapi.Client
	SQLStore *store.SQLStore

	channelArchiverCmd *command.ChannelArchiverCmd
	jobManager         *jobs.JobManager
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case routeRemoveUserFromAllTeamsAndChannels:
		p.handleRemoveUserFromAllTeamsAndChannels(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			fmt.Sprintf("no handler for route %s", r.URL.Path),
		})
	}
}

func (p *Plugin) OnActivate() error {
	p.Client = pluginapi.NewClient(p.API, p.Driver)
	SQLStore, err := store.New(p.Client.Store, &p.Client.Log)
	if err != nil {
		p.Client.Log.Error("cannot create SQLStore", "err", err)
		return err
	}
	p.SQLStore = SQLStore

	// Register slash command for channel archiver
	p.channelArchiverCmd, err = command.RegisterChannelArchiver(p.Client, p.SQLStore)
	if err != nil {
		return fmt.Errorf("cannot register channel archiver slash command: %w", err)
	}

	// Create job manager
	p.jobManager = jobs.NewJobManager(&p.Client.Log)

	// Create job for channel archiver
	channelArchiverJob, err := jobs.NewChannelArchiverJob(ChannelArchiverJobID, p.API, p.Client, SQLStore)
	if err != nil {
		return fmt.Errorf("cannot create channel archiver job: %w", err)
	}
	if err := p.jobManager.AddJob(channelArchiverJob); err != nil {
		return fmt.Errorf("cannot add channel archiver job: %w", err)
	}

	return nil
}

func (p *Plugin) OnDeactivate() error {
	if p.jobManager != nil {
		if err := p.jobManager.Close(time.Second * 15); err != nil {
			return fmt.Errorf("error closing job manager: %w", err)
		}
	}
	return nil
}

func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	cmd, _ := strings.CutPrefix(split[0], "/")

	var response *model.CommandResponse
	var err error

	switch cmd {
	case command.ArchiverTrigger:
		response, err = p.channelArchiverCmd.Execute(args)
	default:
		err = fmt.Errorf("invalid command '%s'", cmd)
	}

	var appErr *model.AppError
	if err != nil {
		appErr = model.NewAppError("", "Error executing command '"+cmd+"'", nil, err.Error(), http.StatusInternalServerError)
	}

	return response, appErr
}
