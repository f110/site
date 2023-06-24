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

func TestLastMod(t *testing.T) {
	d, err := LastMod("testdata/article.md")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2020, 11, 25, 15, 0, 0, 0, time.FixedZone("JST", 9*60*60)).Unix(), d.Unix())

	d, err = LastMod("testdata/article_without_lastmod.md")
	require.NoError(t, err)
	assert.Equal(t, "2020-11-24T18:00:00+09:00", d.Format(time.RFC3339))
}
