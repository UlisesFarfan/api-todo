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
	ColumnCollection    *mongo.Collection
	NoteCollection      *mongo.Collection
	Ctx                 context.Context
}

func NewWorkSpaceRepositoryImpl(Db *mongo.Database, Ctx context.Context) WorkSpacRepository {
	return &WorkSpaceRepositoryImpl{
		UserCollection:      Db.Collection("user"),
		WorkSpaceCollection: Db.Collection("workspace"),
		ColumnCollection:    Db.Collection("column"),
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
	// Delete WorkSpace Columns
	for _, ur := range work_space.Columns {
		filter := bson.M{"_id": ur}
		_, err := d.ColumnCollection.DeleteOne(d.Ctx, filter)
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
func (d *WorkSpaceRepositoryImpl) FindAllByUserId(user_id string) (response.WorkSpaceResponses, error) {
	var user models.User
	workspaces_responses := response.WorkSpaceResponses{}
	oid, _ := primitive.ObjectIDFromHex(user_id)
	filter := bson.M{"_id": oid}
	User := d.UserCollection.FindOne(d.Ctx, filter)
	decodeErr := User.Decode(&user)
	for _, ur := range user.WorkSpaces {
		var workspace_model models.WorkSpace
		filter := bson.M{"_id": ur}
		rol := models.VIEWER
		WorkSpace := d.WorkSpaceCollection.FindOne(d.Ctx, filter)
		_ = WorkSpace.Decode(&workspace_model)
		for _, ur := range workspace_model.Users {
			if ur.UserId.Hex() == user_id {
				rol = ur.Rol
			}
		}
		workspace_response := response.WorkSpaceResponse{
			Id:   workspace_model.Id,
			Name: workspace_model.Name,
			Rol:  rol,
		}
		workspaces_responses = append(workspaces_responses, workspace_response)
	}
	return workspaces_responses, decodeErr
}

// FindById implements WorkSpacRepository.
func (d *WorkSpaceRepositoryImpl) FindById(workSpaceId string, user_id string) (response.WorkSpaceResponseDetail, error) {
	var workspace_model models.WorkSpace
	column_responses := response.ColumnResponses{}
	oid, _ := primitive.ObjectIDFromHex(workSpaceId)
	filter := bson.M{"_id": oid}
	Workspace := d.WorkSpaceCollection.FindOne(d.Ctx, filter)
	decodeErr := Workspace.Decode(&workspace_model)
	rol := models.VIEWER
	for _, ur := range workspace_model.Users {
		if ur.UserId.Hex() == user_id {
			rol = ur.Rol
		}
	}
	for _, i := range workspace_model.Columns {
		var column_model models.Column
		note_responses := response.NoteResponses{}
		filter := bson.M{"_id": i}
		Column := d.ColumnCollection.FindOne(d.Ctx, filter)
		_ = Column.Decode(&column_model)
		for _, j := range column_model.Notes {
			var note_model models.Note
			filter := bson.M{"_id": j}
			Note := d.NoteCollection.FindOne(d.Ctx, filter)
			_ = Note.Decode(&note_model)
			note_response := response.NoteResponse{
				Id:        note_model.Id,
				Task:      note_model.Task,
				CreatedAt: note_model.CreatedAt,
				UpdateAt:  note_model.UpdateAt,
			}
			note_responses = append(note_responses, note_response)
		}
		column_response := response.ColumnResponse{
			Id:        column_model.Id,
			Name:      column_model.Name,
			Notes:     note_responses,
			CreatedAt: column_model.CreatedAt,
			UpdateAt:  column_model.UpdateAt,
		}
		column_responses = append(column_responses, column_response)
	}
	workspace_response := response.WorkSpaceResponseDetail{
		Id:      workspace_model.Id,
		Name:    workspace_model.Name,
		Rol:     rol,
		Columns: column_responses,
	}
	return workspace_response, decodeErr
}

// Save implements WorkSpacRepository.
func (d *WorkSpaceRepositoryImpl) Save(workSpace request.CreateWorkSpaceRequest) error {
	new_workspace := models.WorkSpace{
		Name:      workSpace.Name,
		Columns:   []primitive.ObjectID{},
		CreatedAt: time.Now(),
		UpdateAt:  time.Now(),
	}
	oid, _ := primitive.ObjectIDFromHex(workSpace.UserId)
	user_admin := models.UserRef{
		UserId: oid,
		Rol:    models.OWNER,
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
			Rol:    models.ROL(models.USER),
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
		Id:   work_space.Id,
		Name: work_space.Name,
	}
	return workspace_response, decodeErr
}

// Update implements WorkSpacRepository.
func (d *WorkSpaceRepositoryImpl) UpdateColumnsOrder(workSpace request.UpdateColumnOrderRequest) (response.WorkSpaceResponse, error) {
	var work_space models.WorkSpace
	new_columns_order := []primitive.ObjectID{}
	oid, _ := primitive.ObjectIDFromHex(workSpace.Id)
	filter := bson.M{"_id": oid}
	for _, ur := range workSpace.Columns {
		oid, _ := primitive.ObjectIDFromHex(ur)
		new_columns_order = append(new_columns_order, oid)
	}
	change := bson.M{
		"$set": bson.M{
			"columns": new_columns_order,
		},
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
		Id:   work_space.Id,
		Name: work_space.Name,
	}
	return workspace_response, decodeErr
}
