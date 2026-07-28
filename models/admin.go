package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Admin struct {
	Id       primitive.ObjectID `bson:"_id" json:"id"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
	Username string             `bson:"username" json:"username"`
}

type AdminResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Email       string             `json:"email"`
	Username    string             `bson:"username" json:"username"`
	AccessToken string             `json:"access_token,omitempty"`
}

type SignInAdminInput struct {
	Email    string `json:"email" bson:"email" binding:"required"`
	Password string `json:"password" bson:"password" binding:"required"`
}

func FilteredAdminResponse(admin *Admin) AdminResponse {
	return AdminResponse{
		ID:       admin.Id,
		Email:    admin.Email,
		Username: admin.Username,
	}
}
