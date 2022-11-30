package userz

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	parseConditionRE    = regexp.MustCompile(`^(!?=|>=?|<=?|in|not in|\^|\$)\ ?(.*)$`)
	parseVectorValuesRE = regexp.MustCompile(`\((.*)\)`)
)

// ParseConditionable tries to transform a value in a string condition
// into the specified type T. The following mapping from strings are supported:
//   - string -> string
//   - int -> int
//   - time.RFC3339 -> time.Time
func ParseConditionable[T Conditionable](v string) (any, error) {
	var zero T

	switch t := reflect.TypeOf(zero); t.Kind() {
	case reflect.String:
		return v, nil
	case reflect.Int:
		return strconv.Atoi(v)
	case reflect.Struct:
		if t.Name() == "Time" {
			return time.Parse(time.RFC3339, v)
		}
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind())
	}
}

// ParseCondition tries to parse a string into a condition, matching on it
// and trying to cast the match into the proper Cond[T].
func ParseCondition[T Conditionable](c string) (cond Cond[T], err error) {
	res := parseConditionRE.FindAllStringSubmatch(c, -1)
	if len(res) != 1 || len(res[0]) != 3 {
		err = fmt.Errorf("could not parse condition: %s", c)
		return
	}

	var op Op
	opStr := res[0][1]
	val := res[0][2]
	var value T
	var values []T

	switch opStr {
	case "=": // scalar
		vi, err := ParseConditionable[T](val)
		if err != nil {
			return cond, err
		}

		value = vi.(T)
		op = OpEq
	case "!=": // scalar
		vi, err := ParseConditionable[T](val)
		if err != nil {
			return cond, err
		}

		value = vi.(T)
		op = OpNe
	case ">": // scalar
		vi, err := ParseConditionable[T](val)
		if err != nil {
			return cond, err
		}

		value = vi.(T)
		op = OpGt
	case ">=": // scalar
		vi, err := ParseConditionable[T](val)
		if err != nil {
			return cond, err
		}

		value = vi.(T)
		op = OpGe
	case "<": // scalar
		vi, err := ParseConditionable[T](val)
		if err != nil {
			return cond, err
		}

		value = vi.(T)
		op = OpLt
	case "<=": // scalar
		vi, err := ParseConditionable[T](val)
		if err != nil {
			return cond, err
		}

		value = vi.(T)
		op = OpLe
	case "in": // vector
		vRes := parseVectorValuesRE.FindAllStringSubmatch(val, -1)
		if len(vRes) != 1 || len(vRes[0]) != 2 {
			return cond, fmt.Errorf("malformed vector condition: %s", val)
		}

		for _, el := range strings.Split(vRes[0][1], ",") {
			vi, err := ParseConditionable[T](el)
			if err != nil {
				return cond, err
			}
			values = append(values, vi.(T))
		}

		op = OpInside
	case "not in": // vector
		vRes := parseVectorValuesRE.FindAllStringSubmatch(val, -1)
		if len(vRes) != 1 || len(vRes[0]) != 2 {
			return cond, fmt.Errorf("malformed vector condition: %s", val)
		}

		for _, el := range strings.Split(vRes[0][1], ",") {
			vi, err := ParseConditionable[T](el)
			if err != nil {
				return cond, err
			}
			values = append(values, vi.(T))
		}

		op = OpOutside
	case "^": // scalar
		vi, err := ParseConditionable[T](val)
		if err != nil {
			return cond, err
		}

		value = vi.(T)
		op = OpBegins
	case "$": // scalar
		vi, err := ParseConditionable[T](val)
		if err != nil {
			return cond, err
		}

		value = vi.(T)
		op = OpEnds
	default:
		err = fmt.Errorf("unknown operation: %s", opStr)
		return
	}

	cond.Op = op
	if values != nil {
		cond.Values = values
	} else {
		cond.Value = value
	}

	return
}

// ParseFilter takes an optional map with a set of conditions to be parsed and,
// upon successful parsing of each condition, returns a *Filter.
func ParseFilter(inputMap map[string]string) (*Filter, error) {
	if len(inputMap) == 0 {
		return nil, nil
	}

	var filter Filter

	if firstName, ok := inputMap["first_name"]; ok {
		cond, err := ParseCondition[string](firstName)
		if err != nil {
			return nil, err
		}

		filter.FirstName = cond
	}

	if lastName, ok := inputMap["last_name"]; ok {
		cond, err := ParseCondition[string](lastName)
		if err != nil {
			return nil, err
		}

		filter.LastName = cond
	}

	if nickName, ok := inputMap["nick_name"]; ok {
		cond, err := ParseCondition[string](nickName)
		if err != nil {
			return nil, err
		}

		filter.NickName = cond
	}

	if Email, ok := inputMap["email"]; ok {
		cond, err := ParseCondition[string](Email)
		if err != nil {
			return nil, err
		}

		filter.Email = cond
	}

	if Country, ok := inputMap["country"]; ok {
		cond, err := ParseCondition[string](Country)
		if err != nil {
			return nil, err
		}

		filter.Country = cond
	}

	if CreatedAt, ok := inputMap["created_at"]; ok {
		cond, err := ParseCondition[time.Time](CreatedAt)
		if err != nil {
			return nil, err
		}

		filter.CreatedAt = cond
	}

	if UpdatedAt, ok := inputMap["updated_at"]; ok {
		cond, err := ParseCondition[time.Time](UpdatedAt)
		if err != nil {
			return nil, err
		}

		filter.UpdatedAt = cond
	}

	return &filter, nil
}
