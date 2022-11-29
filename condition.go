package userz

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"time"

	"golang.org/x/exp/constraints"
)

type Conditionable interface {
	constraints.Integer | constraints.Float | string | time.Time
}

type Op int

const (
	OpEq Op = iota
	OpNe
	OpGt
	OpGe
	OpLt
	OpLe
	OpInside
	OpOutside
	OpBegins
	OpEnds
)

func (o Op) String() string {
	switch o {
	case OpEq:
		return "="
	case OpNe:
		return "!="
	case OpGt:
		return ">"
	case OpGe:
		return ">="
	case OpLt:
		return "<"
	case OpLe:
		return "<="
	case OpInside:
		return "∈"
	case OpOutside:
		return "∉"
	case OpBegins:
		return "~^"
	case OpEnds:
		return "~$"
	default:
		panic("Unknown operation")
	}
}

type Cond[T Conditionable] struct {
	Op
	Value  T
	Values []T
}

type ReprCondition[T Conditionable] Cond[T]

var _ Condition[string] = &ReprCondition[string]{}

func (c *ReprCondition[T]) Evaluate(field string) (result any, err error) {
	if err := ValidateOp(c.Op, c.Value, c.Values...); err != nil {
		return "", err
	}

	// apply the operation to the correct type
	switch c.Op {
	case OpEq, OpNe, OpGt, OpGe, OpLt, OpLe, OpBegins, OpEnds: // scalar
		return fmt.Sprintf("%s %s %v", field, c.Op, c.Value), nil
	default: // vector
		values := []T{}
		if c.Values != nil {
			values = c.Values
		}

		return fmt.Sprintf("%s %s %v", field, c.Op, values), nil
	}
}

func (c *ReprCondition[T]) Hash(field string) (string, error) {
	h := fnv.New32a()
	repr, err := c.Evaluate(field)
	if err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(repr.(string))); err != nil {
		return "", err
	}

	return fmt.Sprint(h.Sum32()), nil
}

func ValidateOp[T Conditionable](op Op, value T, values ...T) error {
	var zero T

	if len(values) == 0 { // validation of scalar
		switch t := reflect.TypeOf(zero); t.Kind() {
		case reflect.String:
			if op != OpEq && op != OpNe && op != OpBegins && op != OpEnds {
				return fmt.Errorf("operation not allowed on a string: %s", op)
			}
		case reflect.Struct:
			if t.Name() == "Time" && (op == OpInside || op == OpOutside) && len(values) != 2 {
				return fmt.Errorf("intervals on time.Time must have exactly 2 values, start and end")
			}
			fallthrough
		default:
			if op == OpBegins || op == OpEnds {
				return fmt.Errorf("operation not allowed on a %T: %s", zero, op)
			}
		}
	} else { // validation of vector
		if op != OpInside && op != OpOutside {
			return fmt.Errorf("operation not allowed on a slice of %T: %s", zero, op)
		}
		if reflect.TypeOf(zero).Name() == "Time" && len(values) != 2 {
			return fmt.Errorf("intervals on time.Time must have exactly 2 values, start and end")
		}
	}

	return nil
}
