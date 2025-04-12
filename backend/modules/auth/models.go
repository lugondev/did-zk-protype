package auth

import (
	"time"
)

type User struct {
	ID                 int       `json:"id"`
	Username           string    `json:"username"`
	Email              string    `json:"email,omitempty"`
	PasswordHash       string    `json:"-"`
	DID                string    `json:"did,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	LastLoginAt        time.Time `json:"last_login_at,omitempty"`
	LastUpdatedAt      time.Time `json:"last_updated_at,omitempty"`
	IsTwoFactorEnabled bool      `json:"is_two_factor_enabled"`
}

type Session struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DIDLoginRequest struct {
	DIDID     string `json:"didId"`
	Challenge string `json:"challenge"`
	Signature string `json:"signature"`
	Proof     string `json:"proof"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Email    *string `json:"email,omitempty"`    // Using pointer for optional fields
	Username *string `json:"username,omitempty"` // Using pointer for optional fields
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdateSecurityRequest struct {
	EnableTwoFactor bool `json:"enable_two_factor"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Helper function to create a success response
func NewSuccessResponse(data interface{}, message string) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// Helper function to create an error response
func NewErrorResponse(err error) APIResponse {
	return APIResponse{
		Success: false,
		Error:   err.Error(),
	}
}
