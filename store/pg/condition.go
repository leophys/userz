package pg

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/leophys/userz"
)

const (
	pgTimeFormat = "2006-01-02 15:04:05-07"
)

var _ userz.Condition[string] = &PGCondition[string]{}

type sqlOp[T userz.Conditionable] userz.Op

func (op sqlOp[T]) fmtStr(field string, value T, values ...T) string {
	var valueFmtStr string
	var zero T
	var isTime bool

	switch t := reflect.TypeOf(zero); t.Kind() {
	case reflect.Int, reflect.Uint, reflect.Float32, reflect.Float64:
		valueFmtStr = "%s"
	case reflect.String:
		valueFmtStr = "'%s'"
	case reflect.Struct:
		if t.Name() == "Time" {
			valueFmtStr = "'%s'::TIMESTAMPTZ"
			isTime = true
		}
	}

	switch userz.Op(op) {
	case userz.OpEq:
		return fmt.Sprintf("%s = %s",
			field,
			fmt.Sprintf(valueFmtStr, value),
		)
	case userz.OpNe:
		return fmt.Sprintf("%s != %s",
			field,
			fmt.Sprintf(valueFmtStr, value),
		)
	case userz.OpGt:
		return fmt.Sprintf("%s > %s",
			field,
			fmt.Sprintf(valueFmtStr, value),
		)
	case userz.OpGe:
		return fmt.Sprintf("%s >= %s",
			field,
			fmt.Sprintf(valueFmtStr, value),
		)
	case userz.OpLt:
		return fmt.Sprintf("%s < %s",
			field,
			fmt.Sprintf(valueFmtStr, value),
		)
	case userz.OpLe:
		return fmt.Sprintf("%s <= %s",
			field,
			fmt.Sprintf(valueFmtStr, value),
		)
	case userz.OpInside:
		if isTime {
			return fmt.Sprintf(
				"%s >= %s AND %s <= %s",
				field, fmtTimestamp(valueFmtStr, values[0]),
				field, fmtTimestamp(valueFmtStr, values[1]),
			)
		}

		return fmt.Sprintf("%s IN %s", field, fmtList(valueFmtStr, values...))
	case userz.OpOutside:
		if isTime {
			return fmt.Sprintf(
				"(%s <= %s OR %s >= %s)",
				field, fmtTimestamp(valueFmtStr, values[0]),
				field, fmtTimestamp(valueFmtStr, values[1]),
			)
		}

		return fmt.Sprintf("%s NOT IN %s", field, fmtList(valueFmtStr, values...))
	case userz.OpBegins:
		return fmt.Sprintf("%s LIKE '%s%%'",
			field,
			fmt.Sprint(value),
		)
	case userz.OpEnds:
		return fmt.Sprintf("%s LIKE '%%%s'",
			field,
			fmt.Sprint(value),
		)
	}

	return ""
}

func fmtList[T any](fmtStr string, values ...T) string {
	var fmtValues []string
	for _, v := range values {
		fmtValues = append(fmtValues, fmt.Sprintf(fmtStr, v))
	}

	return fmt.Sprintf("(%s)", strings.Join(fmtValues, ","))
}

func fmtTimestamp(fmtStr string, timestamp any) string {
	ts, ok := timestamp.(time.Time)
	if !ok {
		panic(fmt.Sprintf("not a timestamp: %T", timestamp))
	}

	return fmt.Sprintf(fmtStr, ts.Format(pgTimeFormat))
}

type PGCondition[T userz.Conditionable] userz.Cond[T]

func (c *PGCondition[T]) Evaluate(field string) (any, error) {
	if err := userz.ValidateOp(c.Op, c.Value, c.Values...); err != nil {
		return "", err
	}

	return sqlOp[T](c.Op).fmtStr(field, c.Value, c.Values...), nil
}

func (c *PGCondition[T]) Hash(field string) (string, error) {
	eval, err := c.Evaluate(field)
	if err != nil {
		return "", err
	}

	return hash(eval.(string)), nil
}

func formatFilter(filter *userz.Filter) (string, error) {
	if filter == nil {
		return "1 = 1", nil
	}
	var statements []string

	if filter.Id != "" {
		statements = append(statements, fmt.Sprintf("id = '%s'", filter.Id))
	}

	if filter.FirstName != nil {
		eval, err := filter.FirstName.Evaluate("first_name")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.LastName != nil {
		eval, err := filter.LastName.Evaluate("last_name")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.NickName != nil {
		eval, err := filter.NickName.Evaluate("nickname")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.Email != nil {
		eval, err := filter.Email.Evaluate("email")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.Country != nil {
		eval, err := filter.Country.Evaluate("country")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.CreatedAt != nil {
		eval, err := filter.CreatedAt.Evaluate("created_at")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.UpdatedAt != nil {
		eval, err := filter.UpdatedAt.Evaluate("updated_at")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	return strings.Join(statements, " AND "), nil
}

func hash(filterStr string) string {
	sum := sha256.Sum256([]byte(filterStr))
	return fmt.Sprintf("%x", sum)
}
