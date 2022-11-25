package pg

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"

	"github.com/leophys/userz"
)

var _ userz.Condition[string, string] = &PGCondition[string]{}

type sqlOp[T userz.Conditionable] userz.Op

func (op sqlOp[T]) fmtStr() string {
	var valueFmtStr string
	var zero T
	switch t := reflect.TypeOf(zero); t.Kind() {
	case reflect.Int, reflect.Uint, reflect.Float32, reflect.Float64:
		valueFmtStr = "%s"
	case reflect.String:
		valueFmtStr = "'%s'"
	case reflect.Struct:
		if t.Name() == "Time" {
			valueFmtStr = "'%s'::TIMESTAMPTZ"
		}

	}
	switch userz.Op(op) {
	case userz.OpEq:
		return "%s = " + valueFmtStr
	case userz.OpNe:
		return "%s != " + valueFmtStr
	case userz.OpGt:
		return "%s > " + valueFmtStr
	case userz.OpGe:
		return "%s >= " + valueFmtStr
	case userz.OpLt:
		return "%s < " + valueFmtStr
	case userz.OpLe:
		return "%s <= " + valueFmtStr
	case userz.OpInside:
		return "%s IN " + valueFmtStr
	case userz.OpOutside:
		return "%s NOT IN " + valueFmtStr
	case userz.OpBegins:
		return "%s LIKE '%s%%'"
	case userz.OpEnds:
		return "%s LIKE '%%%s'"
	}

	return ""
}

type PGCondition[T userz.Conditionable] userz.Cond[T]

func (c *PGCondition[T]) Evaluate(field string) (string, error) {
	if err := userz.ValidateOp(c.Op, c.Values...); err != nil {
		return "", err
	}

	fmtStr := sqlOp[T](c.Op).fmtStr()

	// apply the operation to the correct type
	switch c.Op {
	case userz.OpEq, userz.OpNe, userz.OpGt, userz.OpGe, userz.OpLt, userz.OpLe, userz.OpBegins, userz.OpEnds: // scalar
		return fmt.Sprintf(fmtStr, field, c.Value), nil
	default: // vector
		values := []T{}
		if c.Values != nil {
			values = c.Values
		}

		return fmt.Sprintf(fmtStr, field, values), nil
	}
}

func (c *PGCondition[T]) Hash(field string) (string, error) {
	eval, err := c.Evaluate(field)
	if err != nil {
		return "", err
	}

	return hash(eval), nil
}

func formatFilter(filter *userz.Filter[string]) (string, error) {
	var statements []string

	if filter.Id != "" {
		statements = append(statements, fmt.Sprintf("id = '%s'", filter.Id))
	}

	if filter.FirstName != nil {
		eval, err := filter.FirstName.Evaluate("first_name")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval)
	}

	if filter.LastName != nil {
		eval, err := filter.LastName.Evaluate("last_name")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval)
	}

	if filter.NickName != nil {
		eval, err := filter.NickName.Evaluate("nickname")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval)
	}

	if filter.Email != nil {
		eval, err := filter.Email.Evaluate("email")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval)
	}

	if filter.Country != nil {
		eval, err := filter.Country.Evaluate("country")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval)
	}

	if filter.CreatedAt != nil {
		eval, err := filter.CreatedAt.Evaluate("created_at")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval)
	}

	if filter.UpdatedAt != nil {
		eval, err := filter.UpdatedAt.Evaluate("updated_at")
		if err != nil {
			return "", err
		}

		statements = append(statements, eval)
	}

	return strings.Join(statements, " AND "), nil
}

func hash(filterStr string) string {
	sum := sha256.Sum256([]byte(filterStr))
	return fmt.Sprintf("%x", sum)
}
