package channels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/wiggin77/merror"
)

type Reason string

const (
	ReasonDone      Reason = "completed normally"
	ReasonCancelled Reason = "cancelled"
	ReasonTimeout   Reason = "timed out"
	ReasonError     Reason = "error"
)

type ArchiverOpts struct {
	AgeInDays       int
	BatchSize       int
	ExcludeChannels []string // channels names or IDs
	DryRun          bool     // don't archive channels, just count results
}

type ArchiverResults struct {
	ChannelsArchived int
	ExitReason       Reason
	Duration         time.Duration
	Warnings         *merror.MError
	start            time.Time
}

func ArchiveChannels(db *sql.DB, opts ArchiverOpts, ctx context.Context) (results *ArchiverResults, retErr error) {
	results = &ArchiverResults{
		ExitReason: ReasonDone,
		Warnings:   merror.New(),
		start:      time.Now(),
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

	return results, archiveChannels(db, opts, results, ctx)
}

func archiveChannels(db *sql.DB, opts ArchiverOpts, results *ArchiverResults, ctx context.Context) error {
	return errors.New("not implemented yet")
}
