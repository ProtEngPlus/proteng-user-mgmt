package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id        primitive.ObjectID `bson:"_id" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"password"`
	Name      string             `bson:"name" json:"name"`
	Surname   string             `bson:"surname" json:"surname"`
	CitizenId string             `bson:"citizen_id" json:"citizen_id"`
	Role      []string           `bson:"role" json:"role"`
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
	Surname     string             `json:"surname"`
	CitizenId   string             `json:"citizen_id"`
	AccessToken string             `json:"access_token,omitempty"`
	CurrentRole string             `json:"current_role,omitempty"`
}

func FilteredResponse(user *User) UserResponse {
	return UserResponse{
		ID:        user.Id,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		Surname:   user.Surname,
		CitizenId: user.CitizenId,
	}
}
