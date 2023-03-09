package controllers

import (
	"context"
	"reakgo/models"
	"reakgo/utility"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Env struct {
	calendar interface {
		Get(*mongo.Client, context.Context, string, interface{}, interface{}) ([]bson.M, error)
		Put(client *mongo.Client, ctx context.Context, col string, doc interface{}) (*mongo.InsertOneResult, error)
		PostOne(client *mongo.Client, ctx context.Context, col string, filter, update interface{}) (*mongo.UpdateResult, error)
		DeleteOne(client *mongo.Client, ctx context.Context, col string, filter interface{}) (*mongo.DeleteResult, error)
	}
}

var Db *Env

func init() {
	Db = &Env{
		calendar: models.CalanderModel{DB: utility.Db},
	}
}
