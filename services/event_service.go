package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventIn struct {
	Source    string          `json:"source"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type Event struct {
	ID          any            `json:"id" bson:"_id,omitempty"`
	Source      string         `json:"source" bson:"source"`
	Type        string         `json:"type" bson:"type"`
	Timestamp   time.Time      `json:"timestamp" bson:"timestamp"`
	Payload     map[string]any `json:"payload" bson:"payload"`
	PayloadHash string         `json:"-" bson:"payload_hash"` // to manage identical payload
	Count       int64          `json:"count" bson:"count"`
	CreatedAt   time.Time      `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type ListQuery struct {
	Source string
	Type   string
	From   *time.Time
	To     *time.Time
	Limit  int64 // default 50, max 200
	Offset int64 // default 0
}

type EventService interface {
	List(ctx context.Context, query ListQuery) ([]Event, error)
	Upsert(ctx context.Context, in EventIn) (Event, error)
	GetByID(ctx context.Context, id string) (Event, error)
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

func (s *eventService) List(ctx context.Context, query ListQuery) ([]Event, error) {
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	filter := bson.D{}
	if query.Source != "" {
		filter = append(filter, bson.E{Key: "source", Value: query.Source})
	}
	if query.Type != "" {
		filter = append(filter, bson.E{Key: "type", Value: query.Type})
	}
	timeRange := bson.D{}
	if query.From != nil {
		timeRange = append(timeRange, bson.E{Key: "$gte", Value: query.From.UTC()})
	}
	if query.To != nil {
		timeRange = append(timeRange, bson.E{Key: "$lt", Value: query.To.UTC()})
	}
	if len(timeRange) > 0 {
		filter = append(filter, bson.E{Key: "timestamp", Value: timeRange})
	}

	options := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(query.Limit).
		SetSkip(query.Offset)

	cur, err := s.col.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []Event
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *eventService) GetByID(ctx context.Context, id string) (Event, error) {
	var ev Event
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ev, mongo.ErrNoDocuments
	}
	err = s.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&ev)
	return ev, err
}

func (s *eventService) Upsert(ctx context.Context, input EventIn) (Event, error) {
	// Normalize payload into a deterministic byte sequence before hashing.
	var payloadObj map[string]any
	if err := json.Unmarshal(input.Payload, &payloadObj); err != nil {
		return Event{}, fmt.Errorf("parse payload: %w", err)
	}

	canonicalBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return Event{}, fmt.Errorf("canonicalize payload: %w", err)
	}

	sha256Sum := sha256.Sum256(canonicalBytes)
	payloadHash := hex.EncodeToString(sha256Sum[:])

	nowUTC := time.Now().UTC()

	eventTimestamp := input.Timestamp.UTC().Truncate(time.Second)
	if input.Timestamp.IsZero() {
		eventTimestamp = time.Now().UTC().Truncate(time.Second)
	}

	match := bson.M{
		"source":       input.Source,
		"type":         input.Type,
		"payload_hash": payloadHash,
	}
	// $setOnInsert with count=0 + $inc is atomic and guarantees new docs start at 1.
	update := bson.M{
		"$setOnInsert": bson.M{
			"source":       input.Source,
			"type":         input.Type,
			"timestamp":    eventTimestamp,
			"payload":      payloadObj,
			"payload_hash": payloadHash,
			"created_at":   nowUTC,
		},
		"$inc": bson.M{"count": 1},
		"$set": bson.M{"updated_at": nowUTC},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var view Event
	if err := s.col.FindOneAndUpdate(ctx, match, update, opts).Decode(&view); err != nil {
		return Event{}, fmt.Errorf("upsert event: %w", err)
	}
	return view, nil

}
