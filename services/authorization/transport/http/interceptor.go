package auth

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryInterceptor возвращает gRPC-интерсептор, который валидирует JWT access-токен.
func UnaryInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for AuthService methods
		switch info.FullMethod {
		case "/api.v1.AuthService/Register",
			"/api.v1.AuthService/Login",
			"/api.v1.AuthService/Refresh",
			"/api.v1.AuthService/Logout":
			return handler(ctx, req)
		}

		// 1) Извлекаем метаданные
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// 2) Читаем заголовок Authorization
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token not supplied")
		}
		parts := strings.SplitN(authHeaders[0], " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}
		tokenStr := parts[1]

		// 3) Парсим и проверяем JWT
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// 4) Извлекаем user_id из claim “sub”
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "invalid token claims")
		}
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			return nil, status.Error(codes.Unauthenticated, "subject claim (sub) missing in token")
		}

		// 5) Прокидываем user-id в контекст для downstream
		newCtx := context.WithValue(ctx, ctxUserKey{}, sub)

		// 6) Вызываем следующий хендлер
		return handler(newCtx, req)
	}
}
