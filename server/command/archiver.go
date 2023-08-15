package command

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/experimental/command"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-user-deactivation-cleanup/server/channels"
	"github.com/mattermost/mattermost-plugin-user-deactivation-cleanup/server/store"
)

const (
	ArchiverTrigger    = "channel-archiver"
	paramNameDays      = "days"
	paramNameBatchSize = "batch-size"
	paramNameExclude   = "exclude"
	defaultBatchSize   = 50
)

type ErrInvalidSubCommand struct {
	subCommand string
}

func (e ErrInvalidSubCommand) Error() string {
	return "invalid subcommand '" + e.subCommand + "'"
}

type ChannelArchiverCmd struct {
	client   *pluginapi.Client
	sqlStore *store.SQLStore
	commands []*model.AutocompleteData
}

// RegisterChannelArchiver is called by the plugin to register all necessary commands
func RegisterChannelArchiver(client *pluginapi.Client, store *store.SQLStore) (*ChannelArchiverCmd, error) {
	cmdArchive := model.NewAutocompleteData("archive", "", "Archive stale channels")
	cmdList := model.NewAutocompleteData("list", "", "List stale channels that would be archived")
	cmdHelp := model.NewAutocompleteData("help", "", "Display help text")
	commands := []*model.AutocompleteData{cmdArchive, cmdList, cmdHelp}

	cmdArchive.AddNamedTextArgument(paramNameDays, "Number of days of inactivity for a channel to be considered stale", "[int]", "[0-9]*", true)
	cmdArchive.AddNamedTextArgument(paramNameBatchSize, fmt.Sprintf("Channels will be archived in batches of this size. (default=%d)", defaultBatchSize), "[int]", "[0-9]*", false)
	cmdArchive.AddNamedTextArgument(paramNameExclude, "Comma separated list of channel names/IDs to exclude. No Spaces.", "", "", false)

	cmdList.AddNamedTextArgument(paramNameDays, "Number of days of inactivity for a channel to be considered stale", "[int]", "[0-9]*", true)
	cmdList.AddNamedTextArgument(paramNameExclude, "Comma separated list of channel names/IDs to exclude. No Spaces.", "", "", false)

	names := []string{}
	for _, c := range commands {
		names = append(names, c.Trigger)
	}
	hint := "[" + strings.Join(names[:4], "|") + "...]"

	cmd := model.NewAutocompleteData(ArchiverTrigger, hint, "Manage stale channels.")
	cmd.SubCommands = commands

	iconData, err := command.GetIconData(&client.System, "assets/archiver.svg")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get icon data")
	}

	err = client.SlashCommand.Register(&model.Command{
		Trigger:              ArchiverTrigger,
		DisplayName:          "Channel Archiver",
		Description:          "Manage and archive stale channels.",
		AutoComplete:         true,
		AutoCompleteDesc:     strings.Join(names, ", "),
		AutoCompleteHint:     "(subcommand)",
		AutocompleteData:     cmd,
		AutocompleteIconData: iconData,
	})
	if err != nil {
		return nil, err
	}

	return &ChannelArchiverCmd{
		client:   client,
		sqlStore: store,
		commands: commands,
	}, nil
}

func (ca *ChannelArchiverCmd) Execute(args *model.CommandArgs) (*model.CommandResponse, error) {
	params := parseNamedArgs(args.Command)
	subCommand := params[SubCommandKey]

	switch subCommand {
	case "archive":
		return ca.handleArchive(args, params, false)
	case "list":
		return ca.handleArchive(args, params, true)
	case "help":
		return ca.handleHelp()
	default:
		return nil, ErrInvalidSubCommand{subCommand: subCommand}
	}
}

func (ca *ChannelArchiverCmd) handleArchive(args *model.CommandArgs, params map[string]string, list bool) (*model.CommandResponse, error) {
	if !ca.client.User.HasPermissionTo(args.UserId, model.PermissionManageSystem) {
		return responsef("You require %s permissions to execute this command.", model.PermissionManageSystem.Id), nil
	}

	days, err := parseInt(params[paramNameDays], 1, math.MaxInt)
	if err != nil {
		return responsef("Missing or invalid '%s' parameter: %s", paramNameDays, err.Error()), nil
	}

	batchSize := defaultBatchSize
	if bs, ok := params[paramNameBatchSize]; ok {
		batchSize, err = parseInt(bs, 5, 10000)
		if err != nil {
			return responsef("Invalid '%s' parameter: %s", paramNameBatchSize, err.Error()), nil
		}
	}

	var exclude []string
	if ex, ok := params[paramNameExclude]; ok {
		exclude = strings.Split(ex, ",")
	}

	opts := channels.ArchiverOpts{
		AgeInDays:       days,
		BatchSize:       batchSize,
		ExcludeChannels: exclude,
		ListOnly:        list,
	}

	results, err := channels.ArchiveChannels(context.TODO(), ca.sqlStore, ca.client, opts)
	if err != nil {
		return responsef("Error archiving channels: %s", err.Error()), nil
	}

	if list {
		var sb strings.Builder
		for _, ch := range results.ChannelsArchived {
			sb.WriteString(ch)
			sb.WriteString("\n")
		}
		return responsef("%s\nCount: %d\n%s", sb.String(), len(results.ChannelsArchived), results.ExitReason), nil
	}

	return responsef("%d channels archived in %v.\n%d warnings\n%s\n%s",
		len(results.ChannelsArchived), results.Duration, results.Warnings.Len(), results.Warnings.ErrorOrNil(), results.ExitReason), nil
}

func (ca *ChannelArchiverCmd) handleHelp() (*model.CommandResponse, error) {
	resp := ""
	for _, cmd := range ca.commands {
		desc := cmd.Trigger
		if cmd.HelpText != "" {
			desc += " - " + cmd.HelpText
		}
		resp += fmt.Sprintf("/%s %s\n", ArchiverTrigger, desc)
	}

	return responsef(resp), nil
}

// responsef creates an ephemeral command response using printf syntax.
func responsef(format string, args ...interface{}) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf(format, args...),
		Type:         model.PostTypeDefault,
	}
}
