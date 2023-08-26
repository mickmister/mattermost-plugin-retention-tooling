package jobs

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-retention-tooling/server/config"
)

type ChannelArchiverJobSettings struct {
	EnableChannelArchiver bool
	AgeInDays             int
	Frequency             Frequency
	TimeOfDay             time.Time
	ExcludeChannels       []string
	BatchSize             int
}

func (c *ChannelArchiverJobSettings) Clone() *ChannelArchiverJobSettings {
	exclude := make([]string, len(c.ExcludeChannels))
	copy(exclude, c.ExcludeChannels)

	return &ChannelArchiverJobSettings{
		EnableChannelArchiver: c.EnableChannelArchiver,
		AgeInDays:             c.AgeInDays,
		Frequency:             c.Frequency,
		TimeOfDay:             c.TimeOfDay,
		ExcludeChannels:       exclude,
		BatchSize:             c.BatchSize,
	}
}

func (c *ChannelArchiverJobSettings) String() string {
	return fmt.Sprintf("enabled=%T; ageDays=%d; freq=%s; tod=%s; batchSize=%d; excludeLen=%d",
		c.EnableChannelArchiver, c.AgeInDays, c.Frequency, c.TimeOfDay.Format("3:04pm MST"), c.BatchSize, len(c.ExcludeChannels))
}

func parseChannelArchiverJobSettings(cfg *config.Configuration) (*ChannelArchiverJobSettings, error) {
	if !cfg.EnableChannelArchiver {
		return &ChannelArchiverJobSettings{
			EnableChannelArchiver: false,
		}, nil
	}

	if cfg.AgeInDays < config.MinAgeInDays {
		return nil, fmt.Errorf("`Days of inactivity` cannot be less than %d", config.MinAgeInDays)
	}

	freq, err := FreqFromString(cfg.Frequency)
	if err != nil {
		return nil, err
	}

	tod, err := time.Parse("3:04pm MST", cfg.TimeOfDay)
	if err != nil {
		return nil, fmt.Errorf("cannot parse `Time of day`: %w", err)
	}

	nospaces := strings.ReplaceAll(cfg.ExcludeChannels, " ", ",")
	split := strings.Split(nospaces, ",")
	excludes := make([]string, 0)
	for _, s := range split {
		ch := strings.TrimSpace(s)
		if ch != "" {
			excludes = append(excludes, ch)
		}
	}

	if cfg.BatchSize < config.MinBatchSize || cfg.BatchSize > config.MaxBatchSize {
		return nil, fmt.Errorf("`Batch size` cannot be less than %d or more than %d", config.MinBatchSize, config.MaxBatchSize)
	}

	return &ChannelArchiverJobSettings{
		EnableChannelArchiver: cfg.EnableChannelArchiver,
		AgeInDays:             cfg.AgeInDays,
		Frequency:             freq,
		TimeOfDay:             tod,
		ExcludeChannels:       excludes,
	}, nil
}
