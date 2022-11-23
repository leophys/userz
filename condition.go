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

var _ Condition[string, string] = &ReprCondition[string]{}

func (c *ReprCondition[T]) Evaluate() (result string, err error) {
	var zero T

	// validation
	switch t := reflect.TypeOf(zero); t.Kind() {
	case reflect.String:
		if c.Op != OpEq && c.Op != OpNe && c.Op != OpBegins && c.Op != OpEnds {
			return "", fmt.Errorf("operation not allowed on a string: %s", c.Op)
		}
	case reflect.Struct:
		if t.Name() == "Time" && (c.Op == OpInside || c.Op == OpOutside) {
			if len(c.Values) != 2 {
				return "", fmt.Errorf("intervals on time.Time must have exactly 2 values, start and end")
			}
		}
		fallthrough
	default:
		if c.Op == OpBegins || c.Op == OpEnds {
			return "", fmt.Errorf("operation not allowed on a %T: %s", zero, c.Op)
		}
	}

	// apply the operation to the correct type
	switch c.Op {
	case OpEq, OpNe, OpGt, OpGe, OpLt, OpLe, OpBegins, OpEnds: // scalar
		return fmt.Sprintf("%s %v", c.Op, c.Value), nil
	default: // vector
		values := []T{}
		if c.Values != nil {
			values = c.Values
		}

		return fmt.Sprintf("%s %v", c.Op, values), nil
	}
}

func (c *ReprCondition[T]) Hash() (string, error) {
	h := fnv.New32a()
	repr, err := c.Evaluate()
	if err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(repr)); err != nil {
		return "", err
	}

	return fmt.Sprint(h.Sum32()), nil
}
