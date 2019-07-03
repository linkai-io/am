package protocauth

import (
	"context"
	"crypto/subtle"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthFunc func(ctx context.Context) (context.Context, error)

type ServiceAuthFuncOverride interface {
	AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error)
}

func UnaryServerInterceptor(authFunc AuthFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var newCtx context.Context
		var err error
		if overrideSrv, ok := info.Server.(ServiceAuthFuncOverride); ok {
			newCtx, err = overrideSrv.AuthFuncOverride(ctx, info.FullMethod)
		} else {
			newCtx, err = authFunc(ctx)
		}
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// Authenticator is a simple token matching authenticator for grpc/protoc services
type Authenticator struct {
	token []byte
}

func New(token []byte) *Authenticator {
	return &Authenticator{token: token}
}

func (a *Authenticator) Authenticate(ctx context.Context) (context.Context, error) {
	auth, err := extractHeader(ctx, "authorization")
	if err != nil {
		return ctx, err
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ctx, status.Error(codes.Unauthenticated, `missing "Bearer " prefix in "Authorization" header`)
	}

	token := strings.TrimPrefix(auth, prefix)
	if subtle.ConstantTimeCompare([]byte(token), a.token) != 1 {
		return ctx, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Remove token from headers from here on
	return purgeHeader(ctx, "authorization"), nil
}

func extractHeader(ctx context.Context, header string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no headers in request")
	}

	authHeaders, ok := md[header]
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no header in request")
	}

	if len(authHeaders) != 1 {
		return "", status.Error(codes.Unauthenticated, "more than 1 header in request")
	}

	return authHeaders[0], nil
}

func purgeHeader(ctx context.Context, header string) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	mdCopy := md.Copy()
	mdCopy[header] = nil
	return metadata.NewIncomingContext(ctx, mdCopy)
}

type userMDKey struct{}

// UserMetadata contains metadata about a user.
type UserMetadata struct {
	ID string
}

// GetUserMetadata can be used to extract user metadata stored in a context.
func GetUserMetadata(ctx context.Context) (*UserMetadata, bool) {
	userMD := ctx.Value(userMDKey{})

	switch md := userMD.(type) {
	case *UserMetadata:
		return md, true
	default:
		return nil, false
	}
}
