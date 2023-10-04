package column_repository

import (
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ColumnRepositoryImpl struct {
	WorkSpaceCollection *mongo.Collection
	ColumnCollection    *mongo.Collection
	NoteCollection      *mongo.Collection
	Ctx                 context.Context
}

func NewColumnRepositoryImpl(Db *mongo.Database, Ctx context.Context) ColumnRepository {
	return &ColumnRepositoryImpl{
		WorkSpaceCollection: Db.Collection("workspace"),
		ColumnCollection:    Db.Collection("column"),
		NoteCollection:      Db.Collection("note"),
		Ctx:                 Ctx,
	}
}

// Delete implements ColumnRepository.
func (d *ColumnRepositoryImpl) Delete(columnId string) error {
	// Find column to delete notes
	var column_model models.Column
	fmt.Println(columnId)
	oid, _ := primitive.ObjectIDFromHex(columnId)
	filter := bson.M{
		"_id": oid,
	}
	Column := d.ColumnCollection.FindOne(d.Ctx, filter)
	decodeErr := Column.Decode(&column_model)
	if decodeErr != nil {
		return decodeErr
	}
	// For in Notes IDs and Delete
	for _, ur := range column_model.Notes {
		filter := bson.M{
			"_id": ur,
		}
		_, err := d.NoteCollection.DeleteOne(d.Ctx, filter)
		if err != nil {
			return err
		}
	}
	fmt.Println("BBBBB")
	// Delete from WorkSpace
	delete_column_workspace := bson.M{
		"$pull": bson.M{
			"columns": oid,
		},
	}
	filter_column := bson.M{
		"columns": oid,
	}
	result := d.WorkSpaceCollection.FindOneAndUpdate(d.Ctx, filter_column, delete_column_workspace)
	if result.Err() != nil {
		return result.Err()
	}
	// Delete Column
	_, err := d.ColumnCollection.DeleteOne(d.Ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

// FindAll implements ColumnRepository.
func (d *ColumnRepositoryImpl) FindAll(workspaceId string) (response.ColumnResponses, error) {
	panic("unimplemented")
}

// FindById implements ColumnRepository.
func (d *ColumnRepositoryImpl) FindById(columnId string) (response.ColumnResponse, error) {
	var column models.Column
	oid, _ := primitive.ObjectIDFromHex(columnId)
	filter := bson.M{"_id": oid}
	Column := d.ColumnCollection.FindOne(d.Ctx, filter)
	decodeErr := Column.Decode(&column)
	column_response := response.ColumnResponse{
		Id:   column.Id,
		Name: column.Name,
		//Notes:     column.Notes,
		CreatedAt: column.CreatedAt,
		UpdateAt:  column.UpdateAt,
	}
	return column_response, decodeErr
}

// Save implements ColumnRepository.
func (d *ColumnRepositoryImpl) Save(column request.CreateColumnRequest) (response.ColumnResponse, error) {
	new_column := models.Column{
		Name:      column.Name,
		Notes:     []primitive.ObjectID{},
		CreatedAt: time.Now(),
		UpdateAt:  time.Now(),
	}
	oid, _ := primitive.ObjectIDFromHex(column.WorkSpaceId)
	result, err := d.ColumnCollection.InsertOne(d.Ctx, new_column)
	if err != nil {
		return response.ColumnResponse{}, err
	}
	filter := bson.M{"_id": oid}
	after := options.After
	upsert := true
	otp := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	change := bson.M{"$push": bson.M{"columns": result.InsertedID}}
	res := d.WorkSpaceCollection.FindOneAndUpdate(d.Ctx, filter, change, &otp)
	workspace_model := models.WorkSpace{}
	err = res.Decode(&workspace_model)
	if err != nil {
		fmt.Println("ERROOOOR 1", err)
		return response.ColumnResponse{}, err
	}
	filter = bson.M{"_id": result.InsertedID}
	var column_model models.Column
	res_column := d.ColumnCollection.FindOne(d.Ctx, filter)
	err = res_column.Decode(&column_model)
	if err != nil {
		fmt.Println("ERROOOOR 2", err)
		return response.ColumnResponse{}, err
	}
	column_response := response.ColumnResponse{
		Id:        column_model.Id,
		Name:      column_model.Name,
		Notes:     response.NoteResponses{},
		CreatedAt: column_model.CreatedAt,
		UpdateAt:  column_model.UpdateAt,
	}

	return column_response, nil
}

// Update implements ColumnRepository.
func (d *ColumnRepositoryImpl) Update(column request.UpdateColumnRequest) (response.ColumnResponse, error) {
	var column_model models.Column
	oid, _ := primitive.ObjectIDFromHex(column.Id)
	filter := bson.M{"_id": oid}
	update := bson.M{
		"$set": bson.M{
			"name":     column.Name,
			"updateat": time.Now(),
		},
	}
	after := options.After
	upsert := true
	otp := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	result := d.ColumnCollection.FindOneAndUpdate(d.Ctx, filter, update, &otp)
	decodeErr := result.Decode(&column_model)
	column_response := response.ColumnResponse{
		Id:   column_model.Id,
		Name: column_model.Name,
		//Notes:     column_model.Notes,
		CreatedAt: column_model.CreatedAt,
		UpdateAt:  column_model.UpdateAt,
	}
	return column_response, decodeErr
}

func (d *ColumnRepositoryImpl) UpdateNotesOrder(column request.UpdateNotesOrderRequest) (response.ColumnResponse, error) {
	var column_model models.Column
	new_notes_id := []primitive.ObjectID{}
	oid, _ := primitive.ObjectIDFromHex(column.Id)
	for _, ur := range column.Notes {
		oid, _ := primitive.ObjectIDFromHex(ur)
		new_notes_id = append(new_notes_id, oid)
	}
	filter := bson.M{"_id": oid}
	update := bson.M{
		"$set": bson.M{
			"notes":    new_notes_id,
			"updateat": time.Now(),
		},
	}
	after := options.After
	upsert := true
	otp := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	result := d.ColumnCollection.FindOneAndUpdate(d.Ctx, filter, update, &otp)
	decodeErr := result.Decode(&column_model)
	column_response := response.ColumnResponse{
		Id:   column_model.Id,
		Name: column_model.Name,
		//Notes:     column_model.Notes,
		CreatedAt: column_model.CreatedAt,
		UpdateAt:  column_model.UpdateAt,
	}
	return column_response, decodeErr
}
