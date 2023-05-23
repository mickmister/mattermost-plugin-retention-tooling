package main

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) ensureSystemAdmin(userId string) error {
	user, appErr := p.API.GetUser(userId)
	if appErr != nil {
		return errors.Wrapf(appErr, "failed to get user with id %s", userId)
	}

	if !strings.Contains(user.Roles, model.SYSTEM_ADMIN_ROLE_ID) {
		return errors.New("user is not a system admin")
	}

	return nil
}
