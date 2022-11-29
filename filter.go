package userz

import(
	"crypto/sha256"
    "fmt"
)

// Hash returns a unique identifier of the filter.
func (f *Filter) Hash() (string, error) {
	var hashes string

	if f.FirstName != nil {
		hash, err := f.FirstName.Hash("first_name")
		if err != nil {
			return "", fmt.Errorf("failed to get hash for first_name: %w", err)
		}
		hashes = fmt.Sprintf("%s%s", hashes, hash)
	}

	if f.LastName != nil {
		hash, err := f.LastName.Hash("last_name")
		if err != nil {
			return "", fmt.Errorf("failed to get hash for last_name: %w", err)
		}
		hashes = fmt.Sprintf("%s%s", hashes, hash)
	}

	if f.NickName != nil {
		hash, err := f.NickName.Hash("nickname")
		if err != nil {
			return "", fmt.Errorf("failed to get hash for nickname: %w", err)
		}
		hashes = fmt.Sprintf("%s%s", hashes, hash)
	}

	if f.Email != nil {
		hash, err := f.Email.Hash("email")
		if err != nil {
			return "", fmt.Errorf("failed to get hash for email: %w", err)
		}
		hashes = fmt.Sprintf("%s%s", hashes, hash)
	}

	if f.Country != nil {
		hash, err := f.Country.Hash("country")
		if err != nil {
			return "", fmt.Errorf("failed to get hash for country: %w", err)
		}
		hashes = fmt.Sprintf("%s%s", hashes, hash)
	}

	if f.CreatedAt != nil {
		hash, err := f.CreatedAt.Hash("first_name")
		if err != nil {
			return "", fmt.Errorf("failed to get hash for first_name: %w", err)
		}
		hashes = fmt.Sprintf("%s%s", hashes, hash)
	}

	if f.UpdatedAt != nil {
		hash, err := f.UpdatedAt.Hash("updated_at")
		if err != nil {
			return "", fmt.Errorf("failed to get hash for updated_at: %w", err)
		}
		hashes = fmt.Sprintf("%s%s", hashes, hash)
	}

	return Hash(hashes), nil
}

func Hash(str string) string {
	sum := sha256.Sum256([]byte(str))
	return fmt.Sprintf("%x", sum)
}
