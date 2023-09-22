package note_repository

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

type NoteRepositoryImpl struct {
	NoteCollection      *mongo.Collection
	WorkSpaceCollection *mongo.Collection
	Ctx                 context.Context
}

func NewUsersRepositoryImpl(Db *mongo.Database, Ctx context.Context) NoteRepository {
	return &NoteRepositoryImpl{
		NoteCollection:      Db.Collection("note"),
		WorkSpaceCollection: Db.Collection("workspace"),
		Ctx:                 Ctx,
	}
}

// Delete implements NoteRepository.
func (d *NoteRepositoryImpl) Delete(noteId string) error {
	oid, _ := primitive.ObjectIDFromHex(noteId)
	filter := bson.M{
		"_id": oid,
	}
	_, err := d.NoteCollection.DeleteOne(d.Ctx, filter)
	if err != nil {
		return err
	}
	delete_note_workspace := bson.M{
		"$pull": bson.M{
			"notes": oid,
		},
	}
	filter_workspace := bson.M{
		"notes": oid,
	}
	result := d.WorkSpaceCollection.FindOneAndUpdate(d.Ctx, filter_workspace, delete_note_workspace)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

// FindAll implements NoteRepository.
func (d *NoteRepositoryImpl) FindAll() (response.NoteResponses, error) {
	var notes response.NoteResponses
	filter := bson.D{}
	cur, err := d.NoteCollection.Find(d.Ctx, filter)
	if err != nil {
		return nil, err
	}
	for cur.Next(d.Ctx) {
		var note response.NoteResponse
		err := cur.Decode(&note)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

// FindById implements NoteRepository.
func (d *NoteRepositoryImpl) FindById(noteId string) (response.NoteResponse, error) {
	var note response.NoteResponse
	oid, _ := primitive.ObjectIDFromHex(noteId)
	filter := bson.M{"_id": oid}
	Note := d.NoteCollection.FindOne(d.Ctx, filter)
	decodeErr := Note.Decode(&note)
	return note, decodeErr
}

// Save implements NoteRepository.
func (d *NoteRepositoryImpl) Save(note request.CreateNoteRequest, workspaceId string) error {
	new_note := models.Note{
		Task:      note.Task,
		Status:    note.Status,
		CreatedAt: time.Now(),
		UpdateAt:  time.Now(),
	}
	result, err := d.NoteCollection.InsertOne(d.Ctx, new_note)
	if err != nil {
		return err
	}
	oid, _ := primitive.ObjectIDFromHex(workspaceId)
	filter := bson.M{"_id": oid}
	change := bson.M{"$push": bson.M{"notes": result.InsertedID}}
	result_push_workspace := d.WorkSpaceCollection.FindOneAndUpdate(d.Ctx, filter, change)
	workspace_res := bson.D{}
	err = result_push_workspace.Decode(&workspace_res)
	if err != nil {
		return err
	}
	return nil
}

// Update implements NoteRepository.
func (d *NoteRepositoryImpl) Update(notes request.UpdateNoteRequest) (response.NoteResponse, error) {
	var note models.Note
	oid, _ := primitive.ObjectIDFromHex(notes.Id)
	filter := bson.M{"_id": oid}
	update := bson.M{
		"$set": bson.M{
			"task":     notes.Task,
			"status":   notes.Status,
			"updateat": time.Now(),
		},
	}
	after := options.After
	upsert := true
	otp := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	result := d.NoteCollection.FindOneAndUpdate(d.Ctx, filter, update, &otp)
	decodeErr := result.Decode(&note)
	note_response := response.NoteResponse{
		Id:        note.Id,
		Task:      note.Task,
		Status:    note.Status,
		CreatedAt: note.CreatedAt,
		UpdateAt:  note.UpdateAt,
	}
	return note_response, decodeErr
}
