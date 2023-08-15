package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-user-deactivation-cleanup/server/command"
	"github.com/mattermost/mattermost-plugin-user-deactivation-cleanup/server/store"
)

const (
	routeRemoveUserFromAllTeamsAndChannels = "/remove_user_from_all_teams_and_channels"
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
	configuration     *configuration

	Client   *pluginapi.Client
	SQLStore *store.SQLStore

	channelArchiverCmd *command.ChannelArchiverCmd
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

	return nil
}

func (p *Plugin) OnDeactivate() error {
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
