package store

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/mattermost/mattermost-server/v6/model"
)

func (ss *SQLStore) GetStaleChannels(ageInDays int, offset int, batchSize int, excludeChannels []string) ([]*model.Channel, bool, error) {
	olderThan := model.GetMillisForTime(time.Now().AddDate(0, 0, -ageInDays))

	// find all channels where no posts or reactions have been modified,deleted since the olderThan timestamp.
	query := ss.builder.Select("ch.id", "ch.name").Distinct().
		From("channels as ch").
		LeftJoin("posts as p ON ch.id=p.channelid").
		LeftJoin("reactions as r ON p.id=r.postid"). // reactions.channelid does not exist in all versions of server
		Where(sq.Eq{"ch.deleteat": 0}).
		Where(sq.Lt{"ch.updateat": olderThan}).
		Where(sq.Or{sq.Eq{"p.updateat": nil}, sq.Lt{"p.updateat": olderThan, "p.deleteat": olderThan}}).
		Where(sq.Or{sq.Eq{"r.updateat": nil}, sq.Lt{"r.updateat": olderThan, "r.deleteat": olderThan}}).
		OrderBy("ch.id")

	if len(excludeChannels) > 0 {
		query = query.Where(sq.NotEq{"ch.id": excludeChannels, "ch.name": excludeChannels})
	}

	if offset > 0 {
		query = query.Offset(uint64(offset))
	}

	if batchSize > 0 {
		// N+1 to check if there's a next page for pagination
		query = query.Limit(uint64(batchSize) + 1)
	}

	sql, args, _ := query.ToSql()
	fmt.Println(sql, " ::: ", args)

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
