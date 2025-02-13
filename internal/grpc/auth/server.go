package authgrpc

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	Tracer    trace.Tracer
}

func Register(gRPC *grpc.Server, auth Auth, tracer trace.Tracer) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{
		Auth:      auth,
		Validator: NewValidator(),
		Tracer:    tracer,
	})
}

func (s *ServerAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	ctx, span := s.Tracer.Start(ctx, "grpc.server.login")
	defer span.End()

	span.SetAttributes(
		attribute.String("email", req.GetEmail()),
		attribute.Int64("appID", int64(req.GetAppId())),
	)

	err := s.Validator.validateLoginRequest(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		span.RecordError(err)

		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validator error: %s", err))
	}

	token, err := s.Auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		span.RecordError(err)

		return nil, status.Error(codes.Internal, fmt.Sprintf("login error: %s", err))
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *ServerAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	ctx, span := s.Tracer.Start(ctx, "grpc.server.register")
	defer span.End()

	span.SetAttributes(attribute.String("email", req.GetEmail()))

	err := s.Validator.validateRegisterRequest(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validator error: %s", err))
	}

	userID, err := s.Auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("register user error: %s", err))
	}
	span.SetAttributes(attribute.Int64("user_id", userID))

	return &ssov1.RegisterResponse{UsedId: userID}, nil
}

func (s *ServerAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	ctx, span := s.Tracer.Start(ctx, "grpc.server.is_admin")
	defer span.End()

	span.SetAttributes(attribute.Int64("user_id", req.GetUserId()))

	err := s.Validator.validateIsAdminRequest(ctx, req.GetUserId())
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validator error: %s", err))
	}

	isAdmin, err := s.Auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		span.RecordError(err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("is admin error: %s", err))
	}
	span.SetAttributes(attribute.Bool("is_admin", isAdmin))

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}
