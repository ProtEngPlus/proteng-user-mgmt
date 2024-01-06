package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	Email       string             `bson:"email" json:"email"`
	Password    string             `bson:"password" json:"password"`
	Name        string             `bson:"name" json:"name"`
	Surname     string             `bson:"surname" json:"surname"`
	CitizenId   string             `bson:"citizen_id" json:"citizen_id"`
	Role        []string           `bson:"role" json:"role"`
	AccessToken string             `bson:"access_token,omitempty" json:"access_token,omitempty"`
}

type SignInInput struct {
	Email    string `json:"email" bson:"email" binding:"required"`
	Password string `json:"password" bson:"password" binding:"required"`
}
