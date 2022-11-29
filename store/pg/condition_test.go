package pg

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/leophys/userz"
)

var (
	time1Str = "2022-11-23T16:44:26+02:00"
	time2Str = "2022-11-29T00:40:11+02:00"
	time1, _ = time.Parse(time.RFC3339, time1Str)
	time2, _ = time.Parse(time.RFC3339, time2Str)
	time1Exp = "2022-11-23 16:44:26+02"
	time2Exp = "2022-11-29 00:40:11+02"
)

func TestPGCondition_formatFilter(t *testing.T) {
	testCases := []struct {
		filter   *userz.Filter
		expected string
	}{
		{
			filter: &userz.Filter{
				FirstName: &PGCondition[string]{
					Op:    userz.OpEq,
					Value: "john",
				},
			},
			expected: "first_name = 'john'",
		},
		{
			filter: &userz.Filter{
				FirstName: &PGCondition[string]{
					Op:    userz.OpEq,
					Value: "john",
				},
				Country: &PGCondition[string]{
					Op:     userz.OpInside,
					Values: []string{"US", "UK", "CH"},
				},
			},
			expected: "first_name = 'john' AND country IN ('US','UK','CH')",
		},
		{
			filter: &userz.Filter{
				CreatedAt: &PGCondition[time.Time]{
					Op:     userz.OpInside,
					Values: []time.Time{time1, time2},
				},
			},
			expected: fmt.Sprintf(
				"created_at >= '%s'::TIMESTAMPTZ AND created_at <= '%s'::TIMESTAMPTZ",
				time1Exp, time2Exp,
			),
		},
		{
			filter: &userz.Filter{
				CreatedAt: &PGCondition[time.Time]{
					Op:     userz.OpOutside,
					Values: []time.Time{time1, time2},
				},
			},
			expected: fmt.Sprintf(
				"(created_at <= '%s'::TIMESTAMPTZ OR created_at >= '%s'::TIMESTAMPTZ)",
				time1Exp, time2Exp,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			res, err := formatFilter(tc.filter)
			assert.NoError(t, err)
			assert.IsType(t, tc.expected, res)
			assert.Equal(t, tc.expected, res)
		})
	}
}
