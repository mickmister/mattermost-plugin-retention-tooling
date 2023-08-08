package main

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"
)

func (p *Plugin) ensureSystemAdmin(userID string) error {
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return errors.Wrapf(appErr, "failed to get user with id %s", userID)
	}

	if !strings.Contains(user.Roles, model.SystemAdminRoleId) {
		return errors.New("user is not a system admin")
	}

	return nil
}
