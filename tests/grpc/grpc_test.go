//go:build integration

package prototest

import (
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"

	"github.com/leophys/userz"
	"github.com/leophys/userz/internal"
	"github.com/leophys/userz/pkg/proto"
	"github.com/leophys/userz/store/memory"
)

const bufSize = 1 << 20 // 1MB

var listener *bufconn.Listener

func TestGRPCService(t *testing.T) {
	// setup
	assert := assert.New(t)
	require := require.New(t)

	ctx := context.TODO()

	config, err := internal.GetDefaultTLSConfig()
	require.NoError(err)

	store := memory.NewMemoryStore()

	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(config))}
	listener = bufconn.Listen(bufSize)
	s := grpc.NewServer(opts...)
	service := proto.NewUserzServiceServer(store)
	proto.RegisterUserzServer(s, service)

	go func() {
		err := s.Serve(listener)
		require.NoError(err)
	}()

	conn, client, err := dial(ctx)
	require.NoError(err)
	defer conn.Close()

	// expect no user
	list, err := client.List(ctx, &proto.ListRequest{ServiceOrigin: "test", PageSize: 1})
	require.NoError(err)
	listResp, err := list.Recv()
	assert.Error(err)
	assert.Nil(listResp)
	// e, ok := status.FromError(err)
	// require.True(ok)
	// assert.Equal(codes.NotFound, e.Code())

	// add single user
	data1 := &userz.UserData{
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "jd",
		Email:     "jd@morgue.com",
		Password:  "passw0rd",
		Country:   "US",
	}
	add, err := client.Add(ctx, &proto.AddRequest{
		ServiceOrigin: "test",
		Data:          proto.FromUserData(data1),
	})
	require.NoError(err)
	id1 := add.Id
	_, err = uuid.Parse(id1)
	assert.NoError(err)

	// expect the same user
	list, err = client.List(ctx, &proto.ListRequest{ServiceOrigin: "test", PageSize: 1})
	require.NoError(err)
	listResp, err = list.Recv()
	require.NoError(err)
	assert.Len(listResp.Users, 1)
	assert.Equal(id1, listResp.Users[0].Id)
	require.NotNil(listResp.Users[0].FirstName)
	assert.Equal(data1.FirstName, *listResp.Users[0].FirstName)
	require.NotNil(listResp.Users[0].LastName)
	assert.Equal(data1.LastName, *listResp.Users[0].LastName)
	assert.Equal(data1.NickName, listResp.Users[0].NickName)
	assert.Equal(data1.Email, listResp.Users[0].Email)

	require.NotNil(listResp.Users[0].CreatedAt)
	createdAt1, err := time.Parse(time.RFC3339, *listResp.Users[0].CreatedAt)
	require.NoError(err)
	assert.WithinDuration(time.Now(), createdAt1, time.Second)

	assert.Nil(listResp.Users[0].UpdatedAt)

	assert.NoError(
		bcrypt.CompareHashAndPassword(
			[]byte(listResp.Users[0].Password),
			[]byte("passw0rd")),
	)

	listResp, err = list.Recv()
	require.Error(err)
	// e, ok := status.FromError(err)
	// if !ok {
	// 	t.Log(err)
	// }
	// require.True(ok)
	// assert.Equal(codes.NotFound, e.Code())

	// update something
	update, err := client.Update(ctx, &proto.UpdateRequest{
		Id: id1,
		Data: &proto.UserData{
			Email: "test@example.com",
		},
	})
	require.NoError(err)
	assert.Equal(update.User.Email, "test@example.com")

	// add another user
	data2 := &userz.UserData{
		FirstName: "Jane",
		LastName:  "Doe",
		NickName:  "theOne",
		Email:     "one@andonly.com",
		Password:  "iAmTheOneAndOnly",
		Country:   "UK",
	}
	add, err = client.Add(ctx, &proto.AddRequest{
		ServiceOrigin: "test",
		Data:          proto.FromUserData(data2),
	})
	require.NoError(err)
	id2 := add.Id
	_, err = uuid.Parse(id2)
	assert.NoError(err)

	// list users
	list, err = client.List(ctx, &proto.ListRequest{ServiceOrigin: "test", PageSize: 1})
	require.NoError(err)

	listResp, err = list.Recv()
	require.NoError(err)
	assert.Len(listResp.Users, 1)
	assert.Equal(id1, listResp.Users[0].Id)
	require.NotNil(listResp.Users[0].FirstName)
	assert.Equal(data1.FirstName, *listResp.Users[0].FirstName)
	require.NotNil(listResp.Users[0].LastName)
	assert.Equal(data1.LastName, *listResp.Users[0].LastName)
	assert.Equal(data1.NickName, listResp.Users[0].NickName)
	assert.Equal(update.User.Email, listResp.Users[0].Email)

	require.NotNil(listResp.Users[0].CreatedAt)
	createdAt1bis, err := time.Parse(time.RFC3339, *listResp.Users[0].CreatedAt)
	require.NoError(err)
	assert.WithinDuration(createdAt1, createdAt1bis, time.Second)

	require.NotNil(listResp.Users[0].UpdatedAt)
	updatedAt1, err := time.Parse(time.RFC3339, *listResp.Users[0].UpdatedAt)
	require.NoError(err)
	assert.WithinDuration(time.Now(), updatedAt1, 5*time.Second)

	assert.NoError(
		bcrypt.CompareHashAndPassword(
			[]byte(listResp.Users[0].Password),
			[]byte("passw0rd")),
	)

	listResp, err = list.Recv()
	require.NoError(err)
	assert.Len(listResp.Users, 1)
	assert.Equal(id2, listResp.Users[0].Id)
	require.NotNil(listResp.Users[0].FirstName)
	assert.Equal(data2.FirstName, *listResp.Users[0].FirstName)
	require.NotNil(listResp.Users[0].LastName)
	assert.Equal(data2.LastName, *listResp.Users[0].LastName)
	assert.Equal(data2.NickName, listResp.Users[0].NickName)
	assert.Equal(data2.Email, listResp.Users[0].Email)

	require.NotNil(listResp.Users[0].CreatedAt)
	createdAt2, err := time.Parse(time.RFC3339, *listResp.Users[0].CreatedAt)
	require.NoError(err)
	assert.WithinDuration(time.Now(), createdAt2, time.Second)

	assert.Nil(listResp.Users[0].UpdatedAt)

	assert.NoError(
		bcrypt.CompareHashAndPassword(
			[]byte(listResp.Users[0].Password),
			[]byte("iAmTheOneAndOnly")),
	)

	listResp, err = list.Recv()
	require.Error(err)

	// remove user1
	remove, err := client.Remove(ctx, &proto.RemoveRequest{
		ServiceOrigin: "test",
		Id:            id1,
	})
	require.NoError(err)
	assert.Equal(id1, remove.User.Id)

	// check that user is missing from the store
	list, err = client.List(ctx, &proto.ListRequest{ServiceOrigin: "test", PageSize: 1})
	require.NoError(err)

	listResp, err = list.Recv()
	require.NoError(err)
	assert.Len(listResp.Users, 1)
	assert.Equal(id2, listResp.Users[0].Id)

	listResp, err = list.Recv()
	require.Error(err)

}

func dial(ctx context.Context) (*grpc.ClientConn, proto.UserzClient, error) {
	conn, err := grpc.DialContext(ctx, "test",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}),
		),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, err
	}

	return conn, proto.NewUserzClient(conn), nil
}
