package services_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	service "github.com/daadaamed/goeventmanagement/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mustMongo(t *testing.T) (*mongo.Client, *mongo.Database, func()) {
	t.Helper()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	dbName := "events_test_" + time.Now().UTC().Format("20060102_150405_000000")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		cancel()
		t.Skipf("skip: cannot connect to Mongo at %s (%v)", uri, err)
		return nil, nil, func() {}
	}
	if err := client.Ping(ctx, nil); err != nil {
		cancel()
		t.Skipf("skip: cannot ping Mongo at %s (%v)", uri, err)
		return nil, nil, func() {}
	}
	db := client.Database(dbName)
	cleanup := func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
		cancel()
	}
	return client, db, cleanup
}

func TestUpsert_IncrementsCount(t *testing.T) {
	_, db, done := mustMongo(t)
	if db == nil {
		return // skipped
	}
	defer done()

	svc := service.NewEventService(db)

	// 1st upsert -> count=1
	payload1, _ := json.Marshal(map[string]any{"user_id": 1, "ip": "1.2.3.4"})
	ev1, err := svc.Upsert(context.Background(), service.EventIn{
		Source:  "app.web",
		Type:    "user_login",
		Payload: json.RawMessage(payload1),
	})
	if err != nil {
		t.Fatalf("Upsert error: %v", err)
	}
	if ev1.Count != 1 {
		t.Fatalf("expected count=1, got %d", ev1.Count)
	}

	// 2nd upsert (same logical payload with different key order) -> count=2
	payload2 := []byte(`{"ip":"1.2.3.4","user_id":1}`)
	ev2, err := svc.Upsert(context.Background(), service.EventIn{
		Source:  "app.web",
		Type:    "user_login",
		Payload: json.RawMessage(payload2),
	})
	if err != nil {
		t.Fatalf("Upsert error: %v", err)
	}
	if ev2.Count != 2 {
		t.Fatalf("expected count=2, got %d", ev2.Count)
	}
}

func TestGetByID_Roundtrip(t *testing.T) {
	_, db, done := mustMongo(t)
	if db == nil {
		return // skipped
	}
	defer done()

	svc := service.NewEventService(db)
	payload := []byte(`{"a":1}`)
	ev, err := svc.Upsert(context.Background(), service.EventIn{
		Source:  "source.a",
		Type:    "type.a",
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("Upsert error: %v", err)
	}

	// ev.ID is driver-specific type (ObjectID); format to hex if possible
	oid, ok := ev.ID.(interface{ Hex() string })
	if !ok {
		t.Skip("ID is not ObjectID; skipping GetByID test")
	}
	ev2, err := svc.GetByID(context.Background(), oid.Hex())
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if ev2.Source != "source.a" || ev2.Type != "type.a" {
		t.Fatalf("unexpected event: %+v", ev2)
	}
}
