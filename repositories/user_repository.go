package repositories

import (
	"context"
	"errors"

	"github.com/protengplus/proteng-user-mgmt/database"
	"github.com/protengplus/proteng-user-mgmt/utils"

	"github.com/protengplus/proteng-user-mgmt/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	Create(user *models.User) error
	FindById(id string) (*models.User, error)
	Update(id string, user *models.User) error
	Delete(id string) error
	GetAll() ([]*models.User, error)
	FindByEmail(email string) (*models.User, error)
	EnsureIndexes() error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository() UserRepository {
	return &userRepository{collection: database.GetCollection("users")}
}

func (ur *userRepository) GetAll() ([]*models.User, error) {
	var users []*models.User

	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	cursor, err := ur.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
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

	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	var user models.User
	err = ur.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *userRepository) FindByEmail(email string) (*models.User, error) {
	filter := bson.M{"email": utils.NormalizeEmail(email)}

	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	var user models.User
	err := ur.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *userRepository) Create(user *models.User) error {
	user.Id = primitive.NewObjectID()
	user.Email = utils.NormalizeEmail(user.Email)

	hashedPassword, _ := utils.HashPassword(user.Password)
	user.Password = hashedPassword

	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	_, err := ur.collection.InsertOne(ctx, user)
	if err != nil {
		if er, ok := err.(mongo.WriteException); ok && er.WriteErrors[0].Code == 11000 {
			return errors.New("user with that email already exist")
		}
		return err
	}

	// Create a unique index for the email field
	opt := options.Index()
	opt.SetUnique(true)
	index := mongo.IndexModel{Keys: bson.M{"email": 1}, Options: opt}

	if _, err := ur.collection.Indexes().CreateOne(ctx, index); err != nil {
		return errors.New("could not create index for email")
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
			"email":       utils.NormalizeEmail(user.Email),
			"password":    user.Password,
			"name":        user.Name,
			"surname":     user.Surname,
			"role":        user.Role,
			"user_role":   user.UserRole,
			"is_verified": user.IsVerified,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	_, err = ur.collection.UpdateOne(ctx, filter, update)
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

	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	_, err = ur.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

// make sure index on email exists
func (ur *userRepository) EnsureIndexes() error {
	opt := options.Index()
	opt.SetUnique(true)

	// Foo@x.com and FOO@X.COM treated as same
	opt.SetCollation(&options.Collation{
		Locale:   "en",
		Strength: 2,
	})

	index := mongo.IndexModel{Keys: bson.M{"email": 1}, Options: opt}

	if _, err := ur.collection.Indexes().CreateOne(context.Background(), index); err != nil {
		return errors.New("could not create index for email")
	}

	return nil
}
