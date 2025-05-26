package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/NordCoder/Story/services/authorization/entity"

	"github.com/golang-jwt/jwt/v5"
)

type ctxUserKey struct{} // пустая структура для ключа

// UserIDFromCtx извлекает userID из контекста
func UserIDFromCtx(ctx context.Context) (entity.UserID, error) {
	uid, ok := ctx.Value(ctxUserKey{}).(string)
	if !ok || uid == "" {
		return "", entity.ErrUserNotFound
	}
	return entity.UserID(uid), nil
}

// HTTPMiddleware валидирует JWT и кладёт userID в контекст
func HTTPMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Извлечь заголовок
			hdr := r.Header.Get("Authorization")
			parts := strings.SplitN(hdr, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Missing or malformed Authorization header", http.StatusUnauthorized)
				return
			}

			// 2. Проверить JWT
			tokenStr := parts[1]
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid or expired access token", http.StatusUnauthorized)
				return
			}

			// 3. Извлечь userID из claim "sub"
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				http.Error(w, "Subject claim missing in token", http.StatusUnauthorized)
				return
			}

			// 4. Вставить в контекст и пойти дальше
			ctx := context.WithValue(r.Context(), ctxUserKey{}, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
