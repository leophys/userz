package pg

import (
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

func (op sqlOp[T]) fmtStr(field string, value T, values ...T) (string, string) {
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
		outFmt := fmt.Sprintf("%s = ", field) + valueFmtStr
		return outFmt, fmt.Sprintf(outFmt, value)
	case userz.OpNe:
		outFmt := fmt.Sprintf("%s != ", field) + valueFmtStr
		return outFmt, fmt.Sprintf(outFmt, value)
	case userz.OpGt:
		outFmt := fmt.Sprintf("%s > ", field) + valueFmtStr
		return outFmt, fmt.Sprintf(outFmt, value)
	case userz.OpGe:
		outFmt := fmt.Sprintf("%s >= ", field) + valueFmtStr
		return outFmt, fmt.Sprintf(outFmt, value)
	case userz.OpLt:
		outFmt := fmt.Sprintf("%s < ", field) + valueFmtStr
		return outFmt, fmt.Sprintf(outFmt, value)
	case userz.OpLe:
		outFmt := fmt.Sprintf("%s <= ", field) + valueFmtStr
		return outFmt, fmt.Sprintf(outFmt, value)
	case userz.OpInside:
		if isTime {
			outFmt := fmt.Sprintf("%s >= $1 AND %s <= $2", field, field)
			return outFmt, fmt.Sprintf(
				"%s >= %s AND %s <= %s",
				field, fmtTimestamp(valueFmtStr, values[0]),
				field, fmtTimestamp(valueFmtStr, values[1]),
			)
		}

		outFmt := fmt.Sprintf("%s IN ", field) + "%s"
		return outFmt, fmt.Sprintf(outFmt, fmtList(valueFmtStr, values...))
	case userz.OpOutside:
		if isTime {
			outFmt := fmt.Sprintf("(%s <= $1 OR %s >= $2)", field, field)
			return outFmt, fmt.Sprintf(
				"(%s <= %s OR %s >= %s)",
				field, fmtTimestamp(valueFmtStr, values[0]),
				field, fmtTimestamp(valueFmtStr, values[1]),
			)
		}

		outFmt := fmt.Sprintf("%s NOT IN ", field) + "%s"
		return outFmt, fmt.Sprintf(outFmt, fmtList(valueFmtStr, values...))
	case userz.OpBegins:
		outFmt := fmt.Sprintf("%s LIKE ", field) + "'%s%%'"
		return outFmt, fmt.Sprintf(outFmt, fmt.Sprint(value))
	case userz.OpEnds:
		outFmt := fmt.Sprintf("%s LIKE ", field) + "'%%%s'"
		return outFmt, fmt.Sprintf(outFmt, fmt.Sprint(value))
	}

	return "", ""
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

	_, eval := sqlOp[T](c.Op).fmtStr(field, c.Value, c.Values...)
	return eval, nil
}

func (c *PGCondition[T]) Hash(field string) (string, error) {
	if err := userz.ValidateOp(c.Op, c.Value, c.Values...); err != nil {
		return "", err
	}

	fmtStr, _ := sqlOp[T](c.Op).fmtStr(field, c.Value, c.Values...)
	return userz.Hash(fmtStr), nil
}

// NOTE: the type cast of each Condition[T] to *PGCondition[T] is necessary
// to override the implementation of Evaluate.
func formatFilter(filter *userz.Filter) (string, error) {
	if filter == nil {
		return "1 = 1", nil
	}
	var statements []string

	if filter.Id != "" {
		statements = append(statements, fmt.Sprintf("id = '%s'", filter.Id))
	}

	if filter.FirstName != nil {
		eval, err := filter.FirstName.(*PGCondition[string]).Evaluate("first_name")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.LastName != nil {
		eval, err := filter.LastName.(*PGCondition[string]).Evaluate("last_name")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.NickName != nil {
		eval, err := filter.NickName.(*PGCondition[string]).Evaluate("nickname")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.Email != nil {
		eval, err := filter.Email.(*PGCondition[string]).Evaluate("email")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.Country != nil {
		eval, err := filter.Country.(*PGCondition[string]).Evaluate("country")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.CreatedAt != nil {
		eval, err := filter.CreatedAt.(*PGCondition[time.Time]).Evaluate("created_at")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	if filter.UpdatedAt != nil {
		eval, err := filter.UpdatedAt.(*PGCondition[time.Time]).Evaluate("updated_at")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval.(string))
	}

	return strings.Join(statements, " AND "), nil
}
