package repository

import (
	"context"
	"gomongojwt/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	UpdateRefresh(guid string, refresh string) error
	CompareRefreshAndHash(refresh, guid string) (bool, error)
}

type UserRep struct {
	store      *Store
	collection *mongo.Collection
}

func (r *UserRep) UpdateRefresh(guid string, refresh string) error {
	id, err := primitive.ObjectIDFromHex(guid)
	if err != nil {
		return err
	}
	hashToken, err := bcrypt.GenerateFromPassword([]byte(refresh), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := r.collection.UpdateByID(ctx, id, bson.D{{Key: "$set", Value: bson.D{{Key: "refreshtoken", Value: string(hashToken)}}}}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	} else if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
func (r *UserRep) CompareRefreshAndHash(refresh, guid string) (bool, error) {
	id, err := primitive.ObjectIDFromHex(guid)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userRes := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	if userRes.Err() == mongo.ErrNoDocuments {
		return false, mongo.ErrNoDocuments
	}
	usr := &models.User{}
	if err = userRes.Decode(&usr); err != nil {
		return false, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(usr.RefreshToken), []byte(refresh)); err != nil {
		return false, err
	}
	return true, nil
}
