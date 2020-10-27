package middleware

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"new-forum/apiForum/api"
	"strings"
)

func BasicAuth(db *gorm.DB, secret []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email, pass, ok := r.BasicAuth()
			if !ok {
				basicAuthFailed(w)
				return
			}
			user := api.User{}
			result := db.Where("mail = ?", email).First(&user)
			if result.Error != nil {
				basicAuthFailed(w)
				return
			}

			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
			if err != nil {
				basicAuthFailed(w)
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func basicAuthFailed(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, "realm"))
	w.WriteHeader(http.StatusUnauthorized)
}

func TokenAuth(db *gorm.DB, secret []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//Get the JWT on the Authorization Header
			reqToken := r.Header.Get("Authorization")
			splitToken := strings.Split(reqToken, "Bearer ")
			reqToken = splitToken[1]
			token, err := jwt.Parse(reqToken, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(secret), nil
			})
			if err != nil {
				basicAuthFailed(w)
				return
			}
			if !token.Valid {
				basicAuthFailed(w)
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				basicAuthFailed(w)
				return
			}
			if !token.Valid {
				basicAuthFailed(w)
				return
			}
			email, ok := claims["mail"]
			if !ok {
				basicAuthFailed(w)
				return
			}
			user := api.User{}
			result := db.Where("mail = ?", email).First(&user)
			if result.Error != nil {
				basicAuthFailed(w)
				return
			}
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
