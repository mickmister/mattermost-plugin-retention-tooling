package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	papi "github.com/mattermost/mattermost-plugin-api"
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

	papiClient *papi.Client
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
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
	p.papiClient = papi.NewClient(p.API, p.Driver)

	// not implemented yet
	return nil
}

func (p *Plugin) OnDeactivate() error {
	// not implemented yet
	return nil
}
