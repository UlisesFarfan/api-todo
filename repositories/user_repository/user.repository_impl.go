package user_repository

import (
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User services.
type UsersRepositoryImpl struct {
	Collection *mongo.Collection
	Ctx        context.Context
}

func NewUsersRepositoryImpl(Db *mongo.Database, Ctx context.Context) UsersRepository {
	return &UsersRepositoryImpl{Collection: Db.Collection("user"), Ctx: Ctx}
}

// Delete implements UsersRepository.
func (d *UsersRepositoryImpl) Delete(usersId string) error {
	oid, _ := primitive.ObjectIDFromHex(usersId)
	filter := bson.M{
		"_id": oid,
	}
	_, err := d.Collection.DeleteOne(d.Ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

// FindAll implements UsersRepository.
func (d *UsersRepositoryImpl) FindAll() (response.UsersResponses, error) {
	var users response.UsersResponses
	filter := bson.D{}
	cur, err := d.Collection.Find(d.Ctx, filter)
	if err != nil {
		return nil, err
	}
	for cur.Next(d.Ctx) {
		var user response.UsersResponse
		err := cur.Decode(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// FindById implements UsersRepository.
func (d *UsersRepositoryImpl) FindById(userId string) (response.UsersResponse, error) {
	var user models.User
	oid, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.M{"_id": oid}
	User := d.Collection.FindOne(d.Ctx, filter)
	decodeErr := User.Decode(&user)
	user_response := response.UsersResponse{
		Id:        user.Id.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		Img:       user.Img,
		Rol:       string(user.Rol),
		CreatedAt: user.CreatedAt,
		UpdateAt:  user.UpdateAt,
	}
	return user_response, decodeErr
}

// FindByUsername implements UsersRepository.
func (d *UsersRepositoryImpl) FindByEmail(useremail string) (models.User, error) {
	var user models.User
	filter := bson.M{"email": useremail}
	User := d.Collection.FindOne(d.Ctx, filter)
	decodeErr := User.Decode(&user)
	return user, decodeErr
}

// Save implements UsersRepository.
func (d *UsersRepositoryImpl) Save(users request.CreateUsersRequest) error {
	new_user := models.User{
		Name:       users.Name,
		Email:      users.Email,
		Password:   users.Password,
		WorkSpaces: []primitive.ObjectID{},
		Rol:        models.UserRol,
		Img:        users.Img,
		CreatedAt:  time.Now(),
		UpdateAt:   time.Now(),
	}
	_, err := d.Collection.InsertOne(d.Ctx, new_user)
	if err != nil {
		return err
	}
	return nil
}

// Update implements UsersRepository.
func (d *UsersRepositoryImpl) Update(users request.UpdateUsersRequest) (response.UsersResponse, error) {
	var user models.User
	oid, _ := primitive.ObjectIDFromHex(users.Id)
	filter := bson.M{"_id": oid}
	update := bson.M{
		"$set": bson.M{
			"name":     users.Name,
			"email":    users.Email,
			"img":      users.Img,
			"rol":      users.Rol,
			"updateat": time.Now(),
		},
	}
	after := options.After
	upsert := true
	otp := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	result := d.Collection.FindOneAndUpdate(d.Ctx, filter, update, &otp)
	decodeErr := result.Decode(&user)
	user_response := response.UsersResponse{
		Id:        user.Id.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		Img:       user.Img,
		Rol:       string(user.Rol),
		CreatedAt: user.CreatedAt,
		UpdateAt:  user.UpdateAt,
	}
	return user_response, decodeErr
}
