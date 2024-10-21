package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id       primitive.ObjectID `bson:"_id" json:"id"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
	Name     string             `bson:"name" json:"name"`
	Surname  string             `bson:"surname" json:"surname"`
	Role     []string           `bson:"role" json:"role"`
	UserRole string             `bson:"user_role" json:"user_role"`
}

type SignInInput struct {
	Email    string `json:"email" bson:"email" binding:"required"`
	Password string `json:"password" bson:"password" binding:"required"`
	Role     string `bson:"role" json:"role"`
}

type UserResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Email       string             `json:"email"`
	Name        string             `json:"name"`
	Role        []string           `json:"role"`
	UserRole    string             `json:"user_role"`
	Surname     string             `json:"surname"`
	AccessToken string             `json:"access_token,omitempty"`
	CurrentRole string             `json:"current_role,omitempty"`
}

func FilteredResponse(user *User) UserResponse {
	return UserResponse{
		ID:       user.Id,
		Email:    user.Email,
		Name:     user.Name,
		Role:     user.Role,
		UserRole: user.UserRole,
		Surname:  user.Surname,
	}
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordInput struct {
	Password string `json:"password" binding:"required"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}
