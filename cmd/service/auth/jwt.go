package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HollyEllmo/go_rest_tut/cmd/config"
	"github.com/HollyEllmo/go_rest_tut/cmd/types"
	"github.com/HollyEllmo/go_rest_tut/cmd/utils"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserKey contextKey = "userID"

func CreateJWT(secret []byte, userID int) (string, error) {
	expiration := time.Second * time.Duration(config.Envs.JWTExpirationInSeconds)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(userID),
		"expiredAt": time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func WithJWTAuth(handlerFunc http.HandlerFunc, store types.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the token from the user request
		tokenString := getTokenFromRequest(r)
		// validate the JWT
		token, err := validateToken(tokenString)
		if err != nil {
			log.Println("Error validating token:", err)
			permissionDenied(w)
			return
		}

		if !token.Valid {
			log.Println("Invalid token")
			permissionDenied(w)
			return
		}
		// if is we need to fetch the userID from DB (id from the token)
		claims := token.Claims.(jwt.MapClaims)
		str := claims["userID"].(string)

		userID, err := strconv.Atoi(str)
		if err != nil {
			log.Println("Error converting userID to int:", err)
			permissionDenied(w)
			return
		}

		u, err := store.GetUserByID(userID)
		if err != nil {
			log.Println("Error fetching user:", err)
			permissionDenied(w)
			return
		}
		// and set it to the request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, u.ID)
		r = r.WithContext(ctx)

		handlerFunc(w, r)
	}

}

func getTokenFromRequest(r *http.Request) string {
	tokenAuth := r.Header.Get("Authorization")

	if tokenAuth != "" {
		// Проверяем, что заголовок начинается с "Bearer "
		if strings.HasPrefix(tokenAuth, "Bearer ") {
			// Возвращаем токен без префикса "Bearer "
			return strings.TrimPrefix(tokenAuth, "Bearer ")
		}
	}

	return ""
}

func validateToken(t string) (*jwt.Token, error) {
	return jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(config.Envs.JWTSecret), nil
	})
}

func permissionDenied(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}

func GetUserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(UserKey).(int)
	if !ok {
		return -1
	}

	return userID
}
