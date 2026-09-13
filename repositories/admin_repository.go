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

type AdminRepository interface {
	Create(admin *models.Admin) error
	FindById(id string) (*models.Admin, error)
	Update(id string, admin *models.Admin) error
	Delete(id string) error
	GetAll() ([]*models.Admin, error)
	FindByEmail(email string) (*models.Admin, error)
	EnsureIndexes() error
}

type adminRepository struct {
	collection *mongo.Collection
}

func NewAdminRepository() AdminRepository {
	return &adminRepository{collection: database.GetCollection("admins")}
}

func (ur *adminRepository) GetAll() ([]*models.Admin, error) {
	var admins []*models.Admin

	cursor, err := ur.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var admin models.Admin
		if err := cursor.Decode(&admin); err != nil {
			return nil, err
		}
		admins = append(admins, &admin)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return admins, nil
}

func (ur *adminRepository) FindById(id string) (*models.Admin, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectId}

	var admin models.Admin
	err = ur.collection.FindOne(context.Background(), filter).Decode(&admin)
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func (ur *adminRepository) FindByEmail(email string) (*models.Admin, error) {
	filter := bson.M{"email": utils.NormalizeEmail(email)}

	var admin models.Admin
	err := ur.collection.FindOne(context.Background(), filter).Decode(&admin)
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func (ur *adminRepository) Create(admin *models.Admin) error {
	admin.Id = primitive.NewObjectID()
	admin.Email = utils.NormalizeEmail(admin.Email)

	hashedPassword, _ := utils.HashPassword(admin.Password)
	admin.Password = hashedPassword

	_, err := ur.collection.InsertOne(context.Background(), admin)
	if err != nil {
		if er, ok := err.(mongo.WriteException); ok && er.WriteErrors[0].Code == 11000 {
			return errors.New("admin with that email already exist")
		}
		return err
	}

	return nil
}

func (ur *adminRepository) Update(id string, admin *models.Admin) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objectId}

	update := bson.M{
		"$set": bson.M{
			"email":    utils.NormalizeEmail(admin.Email),
			"password": admin.Password,
			"username": admin.Username,
		},
	}

	_, err = ur.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (ur *adminRepository) Delete(id string) error {
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

// make sure index on email exists
func (ur *adminRepository) EnsureIndexes() error {
	opt := options.Index()
	opt.SetUnique(true)
	index := mongo.IndexModel{Keys: bson.M{"email": 1}, Options: opt}

	if _, err := ur.collection.Indexes().CreateOne(context.Background(), index); err != nil {
		return errors.New("could not create index for email")
	}

	return nil
}
