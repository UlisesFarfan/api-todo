package workspace_repository

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

type WorkSpaceRepositoryImpl struct {
	WorkSpaceCollection *mongo.Collection
	UserCollection      *mongo.Collection
	NoteCollection      *mongo.Collection
	Ctx                 context.Context
}

func NewWorkSpaceRepositoryImpl(Db *mongo.Database, Ctx context.Context) WorkSpacRepository {
	return &WorkSpaceRepositoryImpl{
		WorkSpaceCollection: Db.Collection("workspace"),
		UserCollection:      Db.Collection("user"),
		NoteCollection:      Db.Collection("note"),
		Ctx:                 Ctx,
	}
}

// Delete implements WorkSpacRepository.
func (d *WorkSpaceRepositoryImpl) Delete(workSpaceId string) error {
	var work_space models.WorkSpace
	oid, _ := primitive.ObjectIDFromHex(workSpaceId)
	filter := bson.M{"_id": oid}
	Workspace := d.WorkSpaceCollection.FindOne(d.Ctx, filter)
	decodeErr := Workspace.Decode(&work_space)
	if decodeErr != nil {
		return decodeErr
	}
	// Delete WorkSpaceId from Users
	for _, ur := range work_space.Users {
		filter := bson.M{"_id": ur.UserId}
		delete_user_workspace := bson.M{
			"$pull": bson.M{
				"workspaces": oid,
			},
		}
		result := d.UserCollection.FindOneAndUpdate(d.Ctx, filter, delete_user_workspace)
		if result.Err() != nil {
			return result.Err()
		}
	}
	// Delete WorkSpace Notes
	for _, ur := range work_space.Notes {
		filter := bson.M{"_id": ur}
		_, err := d.NoteCollection.DeleteOne(d.Ctx, filter)
		if err != nil {
			return err
		}
	}
	_, err := d.WorkSpaceCollection.DeleteOne(d.Ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

// FindAll implements WorkSpacRepository.
func (d *WorkSpaceRepositoryImpl) FindAll() (response.WorkSpaceResponses, error) {
	panic("unimplemented")
}

// FindById implements WorkSpacRepository.
func (d *WorkSpaceRepositoryImpl) FindById(workSpaceId string) (response.WorkSpaceResponse, error) {
	var work_space models.WorkSpace
	oid, _ := primitive.ObjectIDFromHex(workSpaceId)
	filter := bson.M{"_id": oid}
	Workspace := d.WorkSpaceCollection.FindOne(d.Ctx, filter)
	decodeErr := Workspace.Decode(&work_space)
	workspace_response := response.WorkSpaceResponse{
		Id:        work_space.Id,
		Name:      work_space.Name,
		Users:     work_space.Users,
		Notes:     work_space.Notes,
		CreatedAt: work_space.CreatedAt,
		UpdateAt:  work_space.UpdateAt,
	}
	return workspace_response, decodeErr
}

// Save implements WorkSpacRepository.
func (d *WorkSpaceRepositoryImpl) Save(workSpace request.CreateWorkSpaceRequest) error {
	new_workspace := models.WorkSpace{
		Name:      workSpace.Name,
		Notes:     []primitive.ObjectID{},
		CreatedAt: time.Now(),
		UpdateAt:  time.Now(),
	}
	oid, _ := primitive.ObjectIDFromHex(workSpace.UserId)
	user_admin := models.UserRef{
		UserId: oid,
		Rol:    models.AdminWorkSpace,
	}
	new_workspace.Users = append(new_workspace.Users, user_admin)
	result, err := d.WorkSpaceCollection.InsertOne(d.Ctx, new_workspace)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": oid}
	after := options.After
	upsert := true
	otp := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	change := bson.M{"$push": bson.M{"workspaces": result.InsertedID}}
	res := d.UserCollection.FindOneAndUpdate(d.Ctx, filter, change, &otp)
	user_response := response.UsersResponse{}
	err = res.Decode(&user_response)
	if err != nil {
		return err
	}
	return nil
}

// Update implements WorkSpacRepository.
func (d *WorkSpaceRepositoryImpl) Update(workSpace request.UpdateWorkSpaceRequest) (response.WorkSpaceResponse, error) {
	var work_space models.WorkSpace
	oid, _ := primitive.ObjectIDFromHex(workSpace.Id)
	filter := bson.M{"_id": oid}
	var change bson.M
	if workSpace.UserId != "" {
		oid, _ := primitive.ObjectIDFromHex(workSpace.UserId)
		new_user := models.UserRef{
			UserId: oid,
			Rol:    models.RolWorkSpace(models.UserRol),
		}
		change = bson.M{
			"$push": bson.M{
				"users": new_user,
			},
			"$set": bson.M{
				"updateat": time.Now(),
			},
		}
	} else {
		change = bson.M{
			"$set": bson.M{
				"name":     workSpace.Name,
				"updateat": time.Now(),
			},
		}
	}
	after := options.After
	upsert := true
	otp := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	result := d.WorkSpaceCollection.FindOneAndUpdate(d.Ctx, filter, change, &otp)
	decodeErr := result.Decode(&work_space)
	workspace_response := response.WorkSpaceResponse{
		Id:        work_space.Id,
		Name:      work_space.Name,
		Users:     work_space.Users,
		Notes:     work_space.Notes,
		CreatedAt: work_space.CreatedAt,
		UpdateAt:  work_space.UpdateAt,
	}
	return workspace_response, decodeErr
}
