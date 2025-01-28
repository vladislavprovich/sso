package authgrpc

import (
	"context"
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
	Logout(
		ctx context.Context,
		token string,
	) (*ssov1.LogoutResponse, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth      Auth
	validator validator
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {

	validEmail := s.validator.validateEmail(req.GetEmail())
	if validEmail == false {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}

	validPassword := s.validator.validatePassword(req.GetPassword())
	if validPassword == false {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	if req.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing app ID")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	validEmail := s.validator.validateEmail(req.GetEmail())
	if validEmail == false {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}

	validPassword := s.validator.validatePassword(req.GetPassword())
	if validPassword == false {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}

	return &ssov1.RegisterResponse{UsedId: userID}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	validUserID := s.validator.validateUserID(req.GetUserId())
	if validUserID == false {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	adminBool, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}

	return &ssov1.IsAdminResponse{IsAdmin: adminBool}, nil
}

func (s *serverAPI) Logout(
	ctx context.Context,
	req *ssov1.LogoutRequest,
) (*ssov1.LogoutResponse, error) {
	validToken := s.validator.validateToken(req.GetToken())
	if validToken == false {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	logout, err := s.auth.Logout(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}

	return &ssov1.LogoutResponse{Success: logout.Success}, nil
}
