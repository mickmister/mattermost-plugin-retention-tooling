package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-user-deactivation-cleanup/server/store"
	"github.com/mattermost/mattermost-server/v6/plugin"
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

	return nil
}

func (p *Plugin) OnDeactivate() error {
	// not implemented yet
	return nil
}
