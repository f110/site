package content

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadDateFromFile(t *testing.T) {
	d, err := ReadDateFromFile("testdata/article.md")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2020, 11, 24, 18, 0, 0, 0, time.FixedZone("JST", 9*60*60)).Unix(), d.Unix())
}
