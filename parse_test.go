package userz

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConditionableInt(t *testing.T) {
	assert := assert.New(t)

	v, err := ParseConditionable[int]("1")
	assert.NoError(err)
	_, ok := v.(int)
	assert.True(ok)

	_, err = ParseConditionable[int]("nope")
	assert.Error(err)
}

func TestParseConditionableTime(t *testing.T) {
	assert := assert.New(t)

	now := time.Now().Format(time.RFC3339)

	v, err := ParseConditionable[time.Time](now)
	assert.NoError(err)
	_, ok := v.(time.Time)
	assert.True(ok)
	assert.NotZero(v.(time.Time))

	_, err = ParseConditionable[time.Time]("nope")
	assert.Error(err)
}

func TestParseConditionableString(t *testing.T) {
	assert := assert.New(t)

	v, err := ParseConditionable[string]("of course")
	assert.NoError(err)
	_, ok := v.(string)
	assert.True(ok)
	assert.Equal("of course", v.(string))
}

func TestParseCondition(t *testing.T) {
	testCases := []struct {
		input  string
		err    error
		op     Op
		val    string
		values []string
	}{
		{input: "=ciao", op: OpEq, val: "ciao"},
		{input: "in not_a_list", err: fmt.Errorf("malformed vector condition: not_a_list")},
		{input: "not in (a,b,c)", op: OpOutside, values: []string{"a", "b", "c"}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			cond, err := ParseCondition[string](tc.input)
			if tc.err == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.op, cond.Op)
				assert.Equal(t, tc.val, cond.Value)
				assert.Equal(t, tc.values, cond.Values)
			} else {
				require.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			}
		})
	}
}

func TestParseFilter(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	filter, err := ParseFilter(nil)
	assert.NoError(err)
	assert.Nil(filter)

	filter, err = ParseFilter(map[string]string{})
	assert.NoError(err)
	assert.Nil(filter)

	filter, err = ParseFilter(map[string]string{
		"first_name": "in (John,Jane)",
		"last_name":  "=Doe",
		"created_at": ">2022-11-29T23:55:00+02:00",
	})
	assert.NoError(err)
	require.NotNil(filter)
	assert.Equal(Cond[string]{Op: OpInside, Values: []string{"John", "Jane"}}, filter.FirstName)
	assert.Equal(Cond[string]{Op: OpEq, Value: "Doe"}, filter.LastName)
	limit, err := time.Parse(time.RFC3339, "2022-11-29T23:55:00+02:00")
	require.NoError(err)
	assert.Equal(Cond[time.Time]{Op: OpGt, Value: limit}, filter.CreatedAt)
}
