//go:build integration

package pgintegtest

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/leophys/userz"
	"github.com/leophys/userz/store/pg"
)

var (
	names    = []string{"John", "Jane", "Kane", "Kate", "Tate", "Taty", "Tany", "Vany"}
	surnames = []string{"Brown", "Black", "White", "Green", "Red", "Blue", "Pink", "Rose"}

	nick1 = "nick1"
	nick2 = "nick2"
	nick3 = "nick3"
	nick4 = "nick4"
	nick5 = "nick5"
	nick6 = "nick6"
	nick7 = "nick7"
	nick8 = "nick8"

	password1 = "passw0rd1"
	password2 = "passw0rd2"
	password3 = "passw0rd3"
	password4 = "passw0rd4"
	password5 = "passw0rd5"
	password6 = "passw0rd6"
	password7 = "passw0rd7"
	password8 = "passw0rd8"

	country1 = "UK"
	country2 = "US"

	newUsers = []*userz.UserData{
		newUser(password1, nick1, country1),
		newUser(password2, nick2, country1),
		newUser(password3, nick3, country1),
		newUser(password4, nick4, country1),
		newUser(password5, nick5, country2),
		newUser(password6, nick6, country2),
		newUser(password7, nick7, ""),
		newUser(password8, nick8, ""),
	}
)

func init() {
}

func TestPGStore(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	require.NoError(seed(t))

	ctx := context.TODO()
	dbURL := os.Getenv("POSTGRES_URL")
	skipMigration := os.Getenv("SKIP_MIGRATION")

	// Optionally skip the migration
	if skipMigration == "" {
		err := pg.Migrate(dbURL)
		require.NoError(err)
	}

	// Instantiate the store
	store, err := pg.NewPGStore(ctx, dbURL)
	assert.NoError(err)

	// Add the users
	var users []*userz.User
	usersByCountry := make(map[string][]*userz.User)

	for _, user := range newUsers {
		u, err := store.Add(ctx, user)
		assert.NoError(err)
		assert.Equal(user.FirstName, u.FirstName)
		assert.Equal(user.LastName, u.LastName)
		assert.Equal(user.NickName, u.NickName)
		assert.Equal(user.Email, u.Email)
		assert.WithinDuration(time.Now(), u.CreatedAt, time.Second)

		users = append(users, u)
		switch u.Country {
		case country1:
			usersByCountry[country1] = append(usersByCountry[country1], u)
		case country2:
			usersByCountry[country2] = append(usersByCountry[country2], u)
		default:
			usersByCountry[""] = append(usersByCountry[""], u)
		}
	}

	// List all the users
	it, err := store.List(ctx, nil, 3)
	assert.NoError(err)

	pages := 0
	var usersFound []*userz.User
	for {
		u, err := it.Next(ctx)
		if err != nil {
			if errors.Is(err, userz.ErrNoMorePages) {
				break
			}
			require.Fail("cannot iterate", err)
		}
		pages++
		usersFound = append(usersFound, u...)
	}

	assert.Equal(3, pages)
	assert.Equal(users, usersFound)

	// List only users from country1
	it, err = store.List(ctx, &userz.Filter{Country: &pg.PGCondition[string]{
		Op:    userz.OpEq,
		Value: country1,
	}}, 3)
	assert.NoError(err)

	pages = 0
	usersFound = nil
	for {
		u, err := it.Next(ctx)
		if err != nil {
			if errors.Is(err, userz.ErrNoMorePages) {
				break
			}
			require.Fail("cannot iterate", err)
		}
		pages++
		usersFound = append(usersFound, u...)
	}

	assert.Equal(2, pages)
	assert.Equal(usersByCountry[country1], usersFound)

	// List only users from country1 and whose nick is one of nick4, nick5 or nick6
	it, err = store.List(ctx, &userz.Filter{
		NickName: &pg.PGCondition[string]{
			Op:     userz.OpInside,
			Values: []string{nick4, nick5, nick6},
		},
		Country: &pg.PGCondition[string]{
			Op:    userz.OpEq,
			Value: country1,
		},
	}, 3)
	assert.NoError(err)

	pages = 0
	usersFound = nil
	for {
		u, err := it.Next(ctx)
		if err != nil {
			if errors.Is(err, userz.ErrNoMorePages) {
				break
			}
			require.Fail("cannot iterate", err)
		}
		pages++
		usersFound = append(usersFound, u...)
	}

	assert.Equal(1, pages)
	assert.Len(usersFound, 1)
	assert.Equal(users[3], usersFound[0])

	// Update a user
	u, err := store.Update(ctx, users[0].Id, &userz.UserData{
		Country: "CH",
	})
	assert.NoError(err)
	assert.Equal("CH", u.Country)
	assert.Equal(users[0].FirstName, u.FirstName)
	assert.Equal(users[0].LastName, u.LastName)
	assert.Equal(users[0].NickName, u.NickName)
	assert.Equal(users[0].Email, u.Email)
	assert.Equal(users[0].Password, u.Password)
	assert.Equal(users[0].CreatedAt, u.CreatedAt)
	assert.WithinDuration(time.Now(), u.UpdatedAt, time.Second)

	pageResult, err := store.Page(ctx, &userz.Filter{
		Country: &pg.PGCondition[string]{
			Op:    userz.OpEq,
			Value: "CH",
		},
	}, &userz.PageParams{
		Size:   1,
		Offset: 0,
		Order:  userz.Order{OrdBy: userz.OrdByUpdatedAt, OrdDir: userz.OrdDirAsc},
	})
	assert.NoError(err)
	require.Len(pageResult, 1)
	assert.Equal(u.FirstName, pageResult[0].FirstName)
	assert.Equal(u.LastName, pageResult[0].LastName)
	assert.Equal(u.NickName, pageResult[0].NickName)
	assert.Equal(u.Password, pageResult[0].Password)
	assert.Equal(u.Email, pageResult[0].Email)
	assert.Equal(u.Country, pageResult[0].Country)
	assert.Equal(u.CreatedAt.String(), pageResult[0].CreatedAt.String())
	assert.Equal(u.UpdatedAt.String(), pageResult[0].UpdatedAt.String())

	// Delete a user
	u, err = store.Remove(ctx, users[0].Id)
	assert.NoError(err)
	assert.Equal("CH", u.Country)
	assert.Equal(users[0].FirstName, u.FirstName)
	assert.Equal(users[0].LastName, u.LastName)
	assert.Equal(users[0].NickName, u.NickName)
	assert.Equal(users[0].Email, u.Email)
	assert.Equal(users[0].Password, u.Password)
	assert.Equal(users[0].CreatedAt, u.CreatedAt)

	pageResult, err = store.Page(ctx, &userz.Filter{
		Country: &pg.PGCondition[string]{
			Op:    userz.OpEq,
			Value: "CH",
		},
	}, &userz.PageParams{
		Size:   1,
		Offset: 0,
		Order:  userz.Order{OrdBy: userz.OrdByUpdatedAt, OrdDir: userz.OrdDirAsc},
	})
	assert.NoError(err)
	require.Len(pageResult, 0)
}

func newUser(password, nick, country string) *userz.UserData {
	nidx := rand.Intn(len(names))
	name := names[nidx]
	names = append(names[:nidx], names[nidx+1:]...)

	sidx := rand.Intn(len(surnames))
	surname := surnames[sidx]
	surnames = append(surnames[:sidx], surnames[sidx+1:]...)

	email := strings.ToLower(fmt.Sprintf("%s.%s@band.org", name, surname))

	return &userz.UserData{
		FirstName: name,
		LastName:  surname,
		NickName:  nick,
		Password:  password,
		Email:     email,
		Country:   country,
	}
}

func seed(t *testing.T) (err error) {
	var seed int64

	seedStr := os.Getenv("SEED")
	if seedStr == "" {
		seed = time.Now().UnixNano()
		t.Logf("Using seed: %d", seed)
		rand.Seed(seed)
		return
	}

	seed, err = strconv.ParseInt(seedStr, 10, 64)
	if err != nil {
		return
	}

	t.Logf("Using seed: %d", seed)
	rand.Seed(seed)

	return
}
