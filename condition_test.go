package userz

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateString(t *testing.T) {
	cases := []struct {
		field    string
		cond     *ReprCondition[string]
		expected string
		err      error
	}{
		{field: "test", cond: &ReprCondition[string]{Op: OpEq, Value: "hello"}, expected: `test = hello`},
		{field: "test", cond: &ReprCondition[string]{Op: OpGt, Value: "hello"}, err: fmt.Errorf("operation not allowed on a string: >")},
	}

	for _, tc := range cases {
		var name string
		if tc.err != nil {
			name = fmt.Sprintf("err-%s", tc.err)
		} else {
			name = fmt.Sprintf("success-%s", tc.expected)
		}
		t.Run(name, func(t *testing.T) {
			res, err := tc.cond.Evaluate(tc.field)
			if tc.err != nil {
				require.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}

func TestEvaluateInt(t *testing.T) {
	cases := []struct {
		field    string
		cond     *ReprCondition[int]
		expected string
		err      error
	}{
		{field: "test", cond: &ReprCondition[int]{Op: OpEq, Value: 1}, expected: `test = 1`},
		{field: "test", cond: &ReprCondition[int]{Op: OpInside, Values: []int{1, 2, 3}}, expected: `test ∈ [1 2 3]`},
		{field: "test", cond: &ReprCondition[int]{Op: OpEnds, Value: 1}, err: fmt.Errorf("operation not allowed on a int: ~$")},
	}

	for _, tc := range cases {
		var name string
		if tc.err != nil {
			name = fmt.Sprintf("err-%s", tc.err)
		} else {
			name = fmt.Sprintf("success-%s", tc.expected)
		}
		t.Run(name, func(t *testing.T) {
			res, err := tc.cond.Evaluate(tc.field)
			if tc.err != nil {
				require.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}

func TestEvaluateTime(t *testing.T) {
	testTimeStr := "2022-11-23T16:44:26.371Z"
	testTime, _ := time.Parse(time.RFC3339, testTimeStr)

	cases := []struct {
		field    string
		cond     *ReprCondition[time.Time]
		expected string
		err      error
	}{
		{field: "test_time", cond: &ReprCondition[time.Time]{Op: OpEq, Value: testTime}, expected: fmt.Sprintf("test_time = %s", testTime)},
		{field: "test_time", cond: &ReprCondition[time.Time]{Op: OpOutside}, err: fmt.Errorf("intervals on time.Time must have exactly 2 values, start and end")},
		{field: "test_time", cond: &ReprCondition[time.Time]{Op: OpBegins, Value: testTime}, err: fmt.Errorf("operation not allowed on a time.Time: ~^")},
	}

	for _, tc := range cases {
		var name string
		if tc.err != nil {
			name = fmt.Sprintf("err-%s", tc.err)
		} else {
			name = fmt.Sprintf("success-%s", tc.expected)
		}
		t.Run(name, func(t *testing.T) {
			res, err := tc.cond.Evaluate(tc.field)
			if tc.err != nil {
				require.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}
