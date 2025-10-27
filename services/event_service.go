package service

import (
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventIn struct {
	Source    string          `json:"source"`
	Type      string          `json:"type"`
	Timestamp *time.Time      `json:"timestamp,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type EventView struct {
	ID          any            `json:"id" bson:"_id,omitempty"`
	Source      string         `json:"source" bson:"source"`
	Type        string         `json:"type" bson:"type"`
	Timestamp   time.Time      `json:"timestamp" bson:"timestamp"`
	Payload     map[string]any `json:"payload" bson:"payload"`
	PayloadHash string         `json:"-" bson:"payload_hash"`
	Count       int64          `json:"count" bson:"count"`
	CreatedAt   *time.Time     `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   *time.Time     `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type EventService interface {
	List(ctx context.Context, limit int64) ([]EventView, error)
	GetByID(ctx context.Context, id string) (EventView, error)
}

type eventService struct {
	col *mongo.Collection
}

func NewEventService(db *mongo.Database) EventService {
	col := db.Collection("events")
	// Ensure unique index once.
	_, _ = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "source", Value: 1}, {Key: "type", Value: 1}, {Key: "timestamp", Value: 1}, {Key: "payload_hash", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_src_type_ts_hash"),
	})
	return &eventService{col: col}
}

func (s *eventService) List(ctx context.Context, limit int64) ([]EventView, error) {
	cur, err := s.col.Find(ctx, bson.D{}, options.Find().SetLimit(limit).SetSort(bson.D{{Key: "timestamp", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []EventView
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *eventService) GetByID(ctx context.Context, id string) (EventView, error) {
	var ev EventView
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ev, mongo.ErrNoDocuments
	}
	err = s.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&ev)
	return ev, err
}
