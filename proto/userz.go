package proto

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/leophys/userz"
)

var (
	ErrNoUserFound = status.Error(codes.NotFound, "userz: no user found with given criteria")
	ErrInternal    = status.Error(codes.Internal, "userz: there has been an internal error")
)

type Service struct {
	store userz.Store

	UnimplementedUserzServer
}

func (s *Service) Add(ctx context.Context, req *AddRequest) (*AddResponse, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("origin", req.ServiceOrigin).
		Str("handler", "gRPC-Add").
		Logger()

	raw, err := json.Marshal(req)
	if err != nil {
		logger.Err(err).Msg("Cannot serialize request")
		return nil, ErrInternal
	}
	logger.Debug().
		RawJSON("request", raw).
		Msg("Add request via gRPC")

	user, err := s.store.Add(ctx, req.Data.Into())
	if err != nil {
		logger.Err(err).Msg("Error with the store")
		return nil, ErrInternal
	}
	if user == nil {
		logger.Debug().Msg("No user found")
		return nil, ErrNoUserFound
	}

	return &AddResponse{
		Id: user.Id,
	}, nil
}

func (s *Service) Update(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("origin", req.ServiceOrigin).
		Str("handler", "gRPC-Update").
		Logger()

	raw, err := json.Marshal(req)
	if err != nil {
		logger.Err(err).Msg("Cannot serialize request")
		return nil, ErrInternal
	}
	logger.Debug().
		RawJSON("request", raw).
		Msg("Update request via gRPC")

	user, err := s.store.Update(ctx, req.Id, req.Data.Into())
	if err != nil {
		logger.Err(err).Msg("Error with the store")
		return nil, ErrInternal
	}
	if user == nil {
		logger.Debug().Msg("No user found")
		return nil, ErrNoUserFound
	}

	return &UpdateResponse{
		User: From(user),
	}, nil
}

func (s *Service) Remove(ctx context.Context, req *RemoveRequest) (*RemoveResponse, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("origin", req.ServiceOrigin).
		Str("handler", "gRPC-Remove").
		Logger()

	raw, err := json.Marshal(req)
	if err != nil {
		logger.Err(err).Msg("Cannot serialize request")
		return nil, ErrInternal
	}
	logger.Debug().
		RawJSON("request", raw).
		Msg("Remove request via gRPC")

	user, err := s.store.Remove(ctx, req.Id)
	if err != nil {
		logger.Err(err).Msg("Error with the store")
		return nil, ErrInternal
	}
	if user == nil {
		logger.Debug().Msg("No user found")
		return nil, ErrNoUserFound
	}

	return &RemoveResponse{
		User: From(user),
	}, nil
}

func (s *Service) List(req *ListRequest, server Userz_ListServer) error {
	ctx := server.Context()
	logger := zerolog.Ctx(ctx).
		With().
		Str("origin", req.ServiceOrigin).
		Str("handler", "gRPC-List").
		Logger()

	raw, err := json.Marshal(req)
	if err != nil {
		logger.Err(err).Msg("Cannot serialize request")
		return ErrInternal
	}
	logger.Debug().
		RawJSON("request", raw).
		Msg("List request via gRPC")

	if req.PageSize < 0 {
		return status.Errorf(codes.InvalidArgument, "userz: page_size must be a positive integer")
	}
	pageSize := uint(req.PageSize)

	filter, err := userz.ParseFilter(req.Filter)
	if err != nil {
		logger.Err(err).Msg("Failed to parse filter")
		return status.Errorf(codes.InvalidArgument, "userz: malformed filter")
	}

	iterator, err := s.store.List(ctx, filter, pageSize)
	if err != nil {
		logger.Err(err).Msg("Error with the store")
		return ErrInternal
	}

	for {
		users, err := iterator.Next(ctx)
		if err == userz.ErrNoMorePages {
			break
		}

		var respUsers []*User
		for _, user := range users {
			respUsers = append(respUsers, From(user))
		}

		if err := server.Send(&ListResponse{Users: respUsers}); err != nil {
			logger.Err(err).Msg("Error sending the users' page")
			return err
		}
	}

	return nil
}

func (d *UserData) Into() *userz.UserData {
	firstName := ""
	if d.FirstName != nil {
		firstName = *d.FirstName
	}

	lastName := ""
	if d.LastName != nil {
		lastName = *d.LastName
	}

	country := ""
	if d.Country != nil {
		country = *d.Country
	}

	return &userz.UserData{
		FirstName: firstName,
		LastName:  lastName,
		NickName:  d.NickName,
		Password:  d.Password,
		Email:     d.Email,
		Country:   country,
	}
}

func From(user *userz.User) *User {
	var firstName, lastName, country, createdAt, updatedAt *string

	if user.FirstName != "" {
		firstName = &user.FirstName
	}

	if user.LastName != "" {
		lastName = &user.LastName
	}

	if user.Country != "" {
		country = &user.Country
	}

	if !user.CreatedAt.IsZero() {
		createdAtStr := user.CreatedAt.Format(time.RFC3339)
		createdAt = &createdAtStr
	}

	if !user.UpdatedAt.IsZero() {
		updatedAtStr := user.UpdatedAt.Format(time.RFC3339)
		updatedAt = &updatedAtStr
	}

	return &User{
		FirstName: firstName,
		LastName:  lastName,
		NickName:  user.NickName,
		Email:     user.Email,
		Password:  user.Password.String(),
		Country:   country,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
