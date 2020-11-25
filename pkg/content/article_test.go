package content

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProperty(t *testing.T) {
	t.Run("Date", func(t *testing.T) {
		j := "[[\"‣\",[[\"d\",{\"start_date\":\"2020-11-25\",\"type\":\"date\"}]]]]"
		var data interface{}
		err := json.Unmarshal([]byte(j), &data)
		require.NoError(t, err)

		_ := datePropertyValue(map[string]interface{}{"date": data}, "date")
		assert.Equal(t, "2020-11-25", date.Format("2006-01-02"))
	})
}
