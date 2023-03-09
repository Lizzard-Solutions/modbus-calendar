package models

import (
	"context"
	"os"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Data struct {
	Id          primitive.ObjectID
	FromTime    int64
	ToTime      int64
	Date        string
	Temperature string
	Mode        string
	Hvac        string
	Day         string
}

type CalanderModel struct {
	DB *sqlx.DB
}

func query(client *mongo.Client, ctx context.Context, col string, query interface{}, field interface{}) (*mongo.Cursor, error) {
	// select database and collection.
	collection := client.Database(os.Getenv("DB_NAME")).Collection(col)

	// collection has an method Find, return object of find
	return collection.Find(ctx, query, options.Find())
}

func (auth CalanderModel) Get(client *mongo.Client, ctx context.Context, collection string, filter interface{}, option interface{}) ([]bson.M, error) {
	var results []bson.M // when we use bson.M it should be return value on key value pair
	cursor, err := query(client, ctx, collection, filter, option)
	if err != nil {
		return results, err
	}
	if err := cursor.All(context.TODO(), &results); err != nil { // handle the error
		return results, err
	} else {
		return results, err
	}
}

func (auth CalanderModel) Put(client *mongo.Client, ctx context.Context, col string, doc interface{}) (*mongo.InsertOneResult, error) {
	collection := client.Database(os.Getenv("DB_NAME")).Collection(col)
	// InsertOne accept two argument of type Context
	result, err := collection.InsertOne(ctx, doc)
	return result, err
}

func (auth CalanderModel) PostOne(client *mongo.Client, ctx context.Context, col string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	collection := client.Database(os.Getenv("DB_NAME")).Collection(col)
	// InsertOne accept two argument of type Context
	result, err := collection.UpdateOne(ctx, filter, update)
	return result, err
}

func (auth CalanderModel) DeleteOne(client *mongo.Client, ctx context.Context, col string, filter interface{}) (*mongo.DeleteResult, error) {
	collection := client.Database(os.Getenv("DB_NAME")).Collection(col)
	// InsertOne accept two argument of type Context
	result, err := collection.DeleteOne(ctx, filter)
	return result, err
}
