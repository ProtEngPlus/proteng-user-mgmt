package repositories

import (
	"context"
	"proteng-user-mgmt/database"

	"proteng-user-mgmt/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	Create(user *models.User) error
	FindById(id string) (*models.User, error)
	Update(id string, user *models.User) error
	Delete(id string) error
	GetAll() ([]*models.User, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository() UserRepository {
	return &userRepository{collection: database.GetCollection("users")}
}

func (ur *userRepository) GetAll() ([]*models.User, error) {
	var users []*models.User

	cursor, err := ur.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (ur *userRepository) FindById(id string) (*models.User, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectId}

	var user models.User
	err = ur.collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *userRepository) Create(user *models.User) error {
	user.Id = primitive.NewObjectID()

	_, err := ur.collection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) Update(id string, user *models.User) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objectId}

	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"password":   user.Password,
			"citizen_id": user.CitizenId,
			"name":       user.Name,
			"surname":    user.Surname,
			"role":       user.Role,
		},
	}

	_, err = ur.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) Delete(id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectId}

	_, err = ur.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	return nil
}
