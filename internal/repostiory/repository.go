package repository

import (
	"context"
	"errors"

	"github.com/GroVlAn/doc-store/internal/core"
	"github.com/GroVlAn/doc-store/internal/core/e"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	databaseName       = "documents_db"
	userCollection     = "user"
	tokenCollection    = "token"
	documentCollection = "document"
)

type Repository struct {
	userCollection     *mongo.Collection
	tokenCollection    *mongo.Collection
	documentCollection *mongo.Collection
}

func New(client *mongo.Client) *Repository {
	database := client.Database(databaseName)

	return &Repository{
		userCollection:     database.Collection(userCollection),
		tokenCollection:    database.Collection(tokenCollection),
		documentCollection: database.Collection(documentCollection),
	}
}

func (r *Repository) CreateUser(ctx context.Context, user core.User) error {
	_, err := r.userCollection.InsertOne(ctx, user)
	if err != nil {
		return &e.ErrInsert{Msg: "failed create new user", Err: err}
	}

	return nil
}

func (r *Repository) User(ctx context.Context, login string) (core.User, error) {
	filter := bson.D{{Key: "login", Value: login}}

	var user core.User

	err := r.userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil && err != mongo.ErrNoDocuments {
		return core.User{}, &e.ErrFind{Msg: "user not found", Err: err}
	}
	if err != nil && err == mongo.ErrNoDocuments {
		return core.User{}, e.ErrUserNotFound
	}

	return user, nil
}

func (r *Repository) UserByID(ctx context.Context, id string) (core.User, error) {
	filter := bson.D{{Key: "_id", Value: id}}

	var user core.User

	err := r.userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return core.User{}, &e.ErrFind{Msg: "user not found", Err: err}
	}

	return user, nil
}

func (r *Repository) CreateToken(ctx context.Context, token core.AccessToken) error {
	_, err := r.tokenCollection.InsertOne(ctx, token)
	if err != nil {
		return &e.ErrInsert{Msg: "failed create new token", Err: err}
	}

	return nil
}

func (r *Repository) Token(ctx context.Context, token string) (core.AccessToken, error) {
	filter := bson.D{{Key: "token", Value: token}}

	var accessToken core.AccessToken

	err := r.tokenCollection.FindOne(ctx, filter).Decode(&accessToken)
	if err != nil {
		return core.AccessToken{}, &e.ErrFind{Msg: "token not found", Err: err}
	}

	return accessToken, nil
}

func (r *Repository) DeleteToken(ctx context.Context, token string) error {
	filter := bson.D{{Key: "token", Value: token}}

	_, err := r.tokenCollection.DeleteOne(ctx, filter)
	if err != nil {
		return &e.ErrDelete{Msg: "failed delete token", Err: err}
	}

	return nil
}

func (r *Repository) CreateDocument(ctx context.Context, document core.Document) error {
	_, err := r.documentCollection.InsertOne(ctx, document)
	if err != nil {
		return &e.ErrInsert{Msg: "failed create document", Err: err}
	}

	return nil
}

func (r *Repository) Document(ctx context.Context, login string, documentID string) (core.Document, error) {
	filter := bson.M{
		"_id": documentID,
		"grant": bson.M{
			"$in": []string{login},
		},
	}

	var document core.Document

	err := r.documentCollection.FindOne(ctx, filter).Decode(&document)
	if err != nil {
		return core.Document{}, &e.ErrFind{Msg: "failed to find document", Err: err}
	}

	return document, nil
}

func (r *Repository) DocumentsList(ctx context.Context, df core.DocumentFilter) ([]core.Document, error) {
	filter := bson.M{
		"grant": bson.M{
			"$in": []string{df.Login},
		},
	}

	if len(df.Key) > 0 && len(df.Value) > 0 {
		filter[df.Key] = df.Value
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "Name", Value: -1}, {Key: "created", Value: -1}})
	findOptions.SetLimit(df.Limit)

	cursor, err := r.documentCollection.Find(ctx, filter, findOptions)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, &e.ErrFind{Msg: "failed find documents", Err: err}
	}

	var documentsList []core.Document

	if err := cursor.All(ctx, &documentsList); err != nil {
		return nil, &e.ErrFind{Msg: "failed find documents", Err: err}
	}

	return documentsList, nil
}

func (r *Repository) DocumentByName(ctx context.Context, login string, name string) (core.Document, error) {
	filter := bson.M{
		"grant": bson.M{
			"$in": []string{login},
		},
		"name": name,
	}

	var document core.Document
	err := r.documentCollection.FindOne(ctx, filter).Decode(&document)
	switch {
	case err != nil && errors.Is(err, mongo.ErrNoDocuments):
		return core.Document{}, e.ErrNoDocuments
	case err != nil:
		return core.Document{}, &e.ErrFind{Msg: "failed get document by name", Err: err}
	default:
		return document, nil
	}
}

func (r *Repository) DeleteDocument(ctx context.Context, login string, documentID string) error {
	filter := bson.M{
		"_id": documentID,
		"grant": bson.M{
			"$in": []string{login},
		},
	}

	_, err := r.documentCollection.DeleteOne(ctx, filter)
	if err != nil {
		return &e.ErrDelete{Msg: "failed delete document", Err: err}
	}

	return nil
}
