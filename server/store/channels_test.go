package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLStore_GetStaleChannels(t *testing.T) {
	th := SetupHelper(t).SetupBasic(t)
	defer th.tearDown()

	assert.Nil(t, th.Team1.IsValid())
}
