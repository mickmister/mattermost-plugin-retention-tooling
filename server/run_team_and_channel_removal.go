package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
)

type Payload struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
}

func (p *Plugin) handleRemoveUserFromAllTeamsAndChannels(w http.ResponseWriter, r *http.Request) {
	var writeError = func(errorString string, statusCode int) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(ErrorResponse{errorString})
	}

	if r.Method != http.MethodPost {
		writeError(fmt.Sprintf("unexpected HTTP method %s. Should be POST", r.Method), http.StatusMethodNotAllowed)
		return
	}

	requesterId := r.Header.Get("Mattermost-User-Id")
	if requesterId == "" {
		writeError("request is not from an authenticated user", http.StatusUnauthorized)
		return
	}

	err := p.ensureSystemAdmin(requesterId)
	if err != nil {
		writeError(fmt.Sprintf("error verifying whether user %s is a system admin: %s", requesterId, err.Error()), http.StatusUnauthorized)
		return
	}

	err = p.removeUserFromAllTeamsAndChannels(r, requesterId)
	if err != nil {
		writeError(fmt.Sprintf("error processing request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{true})
}

func (p *Plugin) removeUserFromAllTeamsAndChannels(r *http.Request, requesterId string) error {
	var payload Payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return errors.Wrap(err, "error decoding user info payload")
	}
	r.Body.Close()

	var user *model.User
	var appErr *model.AppError

	// Get user from info supplied in request payload
	if payload.UserId != "" {
		user, appErr = p.API.GetUser(payload.UserId)
		if appErr != nil {
			return errors.Wrapf(appErr, "failed to get user with id %s", payload.UserId)
		}
	} else if payload.Username != "" {
		user, appErr = p.API.GetUserByUsername(payload.Username)
		if appErr != nil {
			return errors.Wrapf(appErr, "failed to get user with username %s", payload.Username)
		}
	} else {
		return errors.New("please provide either user_id or username in the request payload")
	}

	// Start team/channel removal process
	teamMembers, appErr := p.API.GetTeamMembersForUser(user.Id, 0, 1000)
	if appErr != nil {
		return errors.Wrapf(appErr, "failed to get team members for user. user=%s", user.Username)
	}

	for _, tm := range teamMembers {
		err = p.processTeamMember(user, tm.TeamId, requesterId)
		if err != nil {
			return errors.Wrapf(err, "failed to process team member. user=%s team=%s", user.Username, tm.TeamId)
		}
	}

	p.API.LogDebug("Finished for user. user=%s", user.Username)

	return nil
}

func (p *Plugin) processTeamMember(user *model.User, teamId string, requesterId string) error {
	var appErr *model.AppError

	// Remove user from channels in this team
	channelMembers, appErr := p.API.GetChannelMembersForUser(user.Id, teamId, 0, 1000)
	if appErr != nil {
		return errors.Wrapf(appErr, "failed to get channel members")
	}

	for _, cm := range channelMembers {
		err := p.processChannelMember(user, cm.ChannelId)
		if err != nil {
			return errors.Wrapf(err, "failed to process channel member. channel=%s", cm.ChannelId)
		}
	}

	// Remove user from team
	appErr = p.API.DeleteTeamMember(teamId, user.Id, requesterId)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to remove user from team")
	}

	p.API.LogDebug("Removed user from all channels in team. user=%s team=%s", user.Username, teamId)

	return nil
}

func (p *Plugin) processChannelMember(user *model.User, channelId string) error {
	// Remove user from channel
	appErr := p.API.DeleteChannelMember(channelId, user.Id)
	if appErr != nil {
		c, channelErr := p.API.GetChannel(channelId)
		if channelErr != nil {
			return errors.Wrapf(channelErr, "failed to get channel %s", channelId)
		}

		if c.Name == model.DEFAULT_CHANNEL {
			return nil
		}

		return errors.Wrap(appErr, "failed to remove user from channel")
	}

	return nil
}
