package channels

import (
	"context"
	"fmt"
	"time"

	"github.com/wiggin77/merror"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-user-deactivation-cleanup/server/store"
)

type Reason string

const (
	ReasonDone      Reason = "completed normally"
	ReasonCancelled Reason = "canceled"
	ReasonError     Reason = "error"

	sleepBetweenBatchesMillis = 100
)

type ArchiverOpts struct {
	AgeInDays       int
	BatchSize       int
	ExcludeChannels []string // channels names or IDs
	ListOnly        bool     // don't archive channels, just list results
}

type ArchiverResults struct {
	ChannelsArchived []string
	ExitReason       Reason
	Duration         time.Duration
	Warnings         *merror.MError
	start            time.Time
}

func ArchiveChannels(ctx context.Context, store *store.SQLStore, client *pluginapi.Client, opts ArchiverOpts) (results *ArchiverResults, retErr error) {
	results = &ArchiverResults{
		ChannelsArchived: make([]string, 0),
		ExitReason:       ReasonDone,
		Warnings:         merror.New(),
		start:            time.Now(),
	}

	defer func() {
		if p := recover(); p != nil {
			retErr = fmt.Errorf("panic recovered: %v", p)
		}
		if retErr != nil {
			results.ExitReason = ReasonError
		}
		results.Duration = time.Since(results.start)
	}()

	return results, archiveChannels(ctx, store, client, opts, results)
}

func archiveChannels(ctx context.Context, store *store.SQLStore, client *pluginapi.Client, opts ArchiverOpts, results *ArchiverResults) error {
	page := 0
	for {
		select {
		case <-ctx.Done():
			results.ExitReason = ReasonCancelled
			return nil
		default:
		}

		staleChannels, more, err := store.GetStaleChannels(opts.AgeInDays, page*opts.BatchSize, opts.BatchSize, opts.ExcludeChannels)
		if err != nil {
			results.ExitReason = ReasonError
			return fmt.Errorf("cannot fetch stale channels: %w", err)
		}
		page++

		for _, ch := range staleChannels {
			if opts.ListOnly {
				results.ChannelsArchived = append(results.ChannelsArchived, fmt.Sprintf("%s (%s)", ch.Id, ch.Name))
				continue
			}

			// archive the channel; record errors but keep going.
			appErr := client.Channel.Delete(ch.Id)
			if appErr != nil {
				results.Warnings.Append(fmt.Errorf("could not archive channel %s (%s): %w", ch.Id, ch.Name, appErr))
				continue
			}
			results.ChannelsArchived = append(results.ChannelsArchived, fmt.Sprintf("%s (%s)", ch.Id, ch.Name))

			// sleep a short time so we don't peg the cpu
			select {
			case <-time.After(time.Millisecond * sleepBetweenBatchesMillis):
			case <-ctx.Done():
				results.ExitReason = ReasonCancelled
				return nil
			}
		}

		if !more {
			break
		}

		// sleep a short time so we don't peg the cpu
		select {
		case <-time.After(time.Millisecond * sleepBetweenBatchesMillis):
		case <-ctx.Done():
			results.ExitReason = ReasonCancelled
			return nil
		}
	}

	return nil
}
