package pg

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/leophys/userz"
)

var _ userz.Condition[string, string] = &PGCondition[string]{}

type PGCondition[T userz.Conditionable] struct{}

func (c *PGCondition[T]) Evaluate() (string, error)
func (c *PGCondition[T]) Hash() (string, error)

func formatFilter(filter *userz.Filter[string]) (string, error) {
	var statements []string

	if filter.Id != "" {
		statements = append(statements, fmt.Sprintf("id='%s'", filter.Id))
	}

	if filter.FirstName != nil {
		eval, err := filter.FirstName.Evaluate()
		if err != nil {
			return "", err
		}

		statements = append(statements, fmt.Sprintf("first_name %s", eval))
	}

	if filter.LastName != nil {
		eval, err := filter.LastName.Evaluate()
		if err != nil {
			return "", err
		}

		statements = append(statements, fmt.Sprintf("last_name %s", eval))
	}

	if filter.NickName != nil {
		eval, err := filter.NickName.Evaluate()
		if err != nil {
			return "", err
		}

		statements = append(statements, fmt.Sprintf("nick_name %s", eval))
	}

	if filter.Email != nil {
		eval, err := filter.Email.Evaluate()
		if err != nil {
			return "", err
		}

		statements = append(statements, fmt.Sprintf("email %s", eval))
	}

	if filter.Country != nil {
		eval, err := filter.Country.Evaluate()
		if err != nil {
			return "", err
		}

		statements = append(statements, fmt.Sprintf("country %s", eval))
	}

	if filter.CreatedAt != nil {
		eval, err := filter.CreatedAt.Evaluate()
		if err != nil {
			return "", err
		}

		statements = append(statements, fmt.Sprintf("created_at %s", eval))
	}

	if filter.UpdatedAt != nil {
		eval, err := filter.UpdatedAt.Evaluate()
		if err != nil {
			return "", err
		}

		statements = append(statements, fmt.Sprintf("updated_at %s", eval))
	}

	return strings.Join(statements, " AND "), nil
}

func hash(filterStr string) string {
	sum := sha256.Sum256([]byte(filterStr))
	return fmt.Sprintf("%x", sum)
}
