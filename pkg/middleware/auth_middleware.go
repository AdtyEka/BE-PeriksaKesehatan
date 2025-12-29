package middleware

import (
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/pkg/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey = "userID"

// AuthMiddleware membuat middleware untuk validasi JWT token
func AuthMiddleware(authRepo *repository.AuthRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil token dari header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
			c.Abort()
			return
		}

		// Parse token (format: "Bearer <token>")
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		// Cek apakah token sudah di-blacklist
		isBlacklisted, err := authRepo.IsTokenBlacklisted(tokenString)
		if err != nil {
			utils.InternalServerError(c, "Gagal memeriksa status token", err.Error())
			c.Abort()
			return
		}
		if isBlacklisted {
			utils.Unauthorized(c, "Token tidak valid atau sudah expired")
			c.Abort()
			return
		}

		// Parse dan validasi token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validasi bahwa signing method adalah HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
			c.Abort()
			return
		}

		if !token.Valid {
			utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
			c.Abort()
			return
		}

		// Ambil claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
			c.Abort()
			return
		}

		// Validasi expiry time
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				utils.Unauthorized(c, "Token tidak valid atau sudah expired")
				c.Abort()
				return
			}
		}

		// Ambil user ID dari claims
		sub, ok := claims["sub"]
		if !ok {
			utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
			c.Abort()
			return
		}

		// Convert ke uint
		var userID uint
		if userIDFloat, ok := sub.(float64); ok {
			userID = uint(userIDFloat)
		} else if userIDStr, ok := sub.(string); ok {
			userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
				c.Abort()
				return
			}
			userID = uint(userIDUint)
		} else {
			utils.Unauthorized(c, "Token tidak valid atau tidak ditemukan")
			c.Abort()
			return
		}

		// Set userID ke context untuk digunakan di handler
		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// GetUserIDFromContext mengambil user ID dari gin context
// Harus dipanggil setelah AuthMiddleware
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, false
	}

	return id, true
}

