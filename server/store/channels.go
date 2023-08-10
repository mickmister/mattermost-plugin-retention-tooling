package store

import (
	sq "github.com/Masterminds/squirrel"

	"github.com/mattermost/mattermost-server/v6/model"
)

const (
	MillisPerDay = 86400000
)

func (ss *SQLStore) GetStaleChannels(ageInDays int, offset int, batchSize int, excludeChannels []string) ([]*model.Channel, bool, error) {
	olderThan := model.GetMillis() - (int64(ageInDays) * MillisPerDay)

	// find all channels where no posts or reactions have been created,modified,deleted since the olderThan timestamp.
	query := ss.builder.Select("ch.id", "ch.name").
		From("channels as ch").
		Join("posts as p ON ch.id=p.channelid").
		Join("reactions as r ON ch.id=r.channelid").
		Where(sq.Eq{"ch.deleteat": 0}).
		Where(sq.NotEq{"ch.id": excludeChannels}).
		Where(sq.NotEq{"ch.name": excludeChannels}).
		Where(sq.Lt{"p.updateat": olderThan, "p.deleteat": olderThan}).
		Where(sq.Lt{"r.updateat": olderThan, "r.deleteat": olderThan}).
		OrderBy("ch.id")

	if offset > 0 {
		query = query.Offset(uint64(offset))
	}

	if batchSize > 0 {
		// N+1 to check if there's a next page for pagination
		query = query.Limit(uint64(batchSize) + 1)
	}

	rows, err := query.Query()
	if err != nil {
		ss.logger.Error("error fetching stale channels", "err", err)
		return nil, false, err
	}

	channels := []*model.Channel{}
	for rows.Next() {
		channel := &model.Channel{}

		if err := rows.Scan(&channel.Id, &channel.Name); err != nil {
			ss.logger.Error("error scanning stale channels", "err", err)
			return nil, false, err
		}
		channels = append(channels, channel)
	}

	var hasMore bool
	if batchSize > 0 && len(channels) > batchSize {
		hasMore = true
		channels = channels[0:batchSize]
	}

	return channels, hasMore, nil
}
