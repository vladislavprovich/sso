package authgrpc

import (
	"context"
	"fmt"

	ssov1 "github.com/vladislavprovich/protobufContract/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type ServerAPI struct {
	ssov1.UnimplementedAuthServer
	Auth      Auth
	Validator *Validator
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{
		Auth:      auth,
		Validator: NewValidator(),
	})
}

func (s *ServerAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	err := s.Validator.validateLoginRequest(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validator error: %s", err))
	}

	token, err := s.Auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("login error: %s", err))
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *ServerAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	err := s.Validator.validateRegisterRequest(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validator error: %s", err))
	}

	userID, err := s.Auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("register user error: %s", err))
	}

	return &ssov1.RegisterResponse{UsedId: userID}, nil
}

func (s *ServerAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	err := s.Validator.validateIsAdminRequest(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validator error: %s", err))
	}

	isAdmin, err := s.Auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("is admin error: %s", err))
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}
