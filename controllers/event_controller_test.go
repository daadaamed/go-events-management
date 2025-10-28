package controllers_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daadaamed/goeventmanagement/controllers"
	service "github.com/daadaamed/goeventmanagement/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// --- mock service ---
type mockSvc struct {
	listFn   func() ([]service.Event, error)
	upsertFn func(in service.EventIn) (service.Event, error)
	getByID  func(id string) (service.Event, error)
}

func (m *mockSvc) List(_ context.Context, _ service.ListQuery) ([]service.Event, error) {
	return m.listFn()
}
func (m *mockSvc) Upsert(_ context.Context, in service.EventIn) (service.Event, error) {
	return m.upsertFn(in)
}
func (m *mockSvc) GetByID(_ context.Context, id string) (service.Event, error) {
	return m.getByID(id)
}

func newRouter(h *controllers.EventHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.RegisterRoutes(r)
	return r
}

func TestGetEvents_OK(t *testing.T) {
	s := &mockSvc{
		listFn: func() ([]service.Event, error) {
			return []service.Event{
				{Source: "app.web", Type: "login", Count: 2},
			}, nil
		},
		upsertFn: func(in service.EventIn) (service.Event, error) { return service.Event{}, nil },
		getByID:  func(id string) (service.Event, error) { return service.Event{}, nil },
	}
	h := controllers.NewEventHandler(s)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	if got := w.Body.String(); !bytes.Contains([]byte(got), []byte(`"count":2`)) {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestGetEventByID_NotFound(t *testing.T) {
	s := &mockSvc{
		listFn: func() ([]service.Event, error) { return nil, nil },
		upsertFn: func(in service.EventIn) (service.Event, error) {
			return service.Event{}, nil
		},
		getByID: func(id string) (service.Event, error) {
			return service.Event{}, mongo.ErrNoDocuments
		},
	}
	h := controllers.NewEventHandler(s)
	r := gin.New()
	r.GET("/events/:id", h.GetEventByID)

	req := httptest.NewRequest(http.MethodGet, "/events/64f000000000000000000000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestPostEvent_Validation(t *testing.T) {
	s := &mockSvc{
		listFn:   func() ([]service.Event, error) { return nil, nil },
		upsertFn: func(in service.EventIn) (service.Event, error) { return service.Event{}, nil },
		getByID:  func(id string) (service.Event, error) { return service.Event{}, nil },
	}
	h := controllers.NewEventHandler(s)
	r := newRouter(h)

	// missing fields -> 400
	body := bytes.NewBufferString(`{"payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/events", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPostEvent_OK(t *testing.T) {
	now := time.Now().UTC()
	s := &mockSvc{
		listFn: func() ([]service.Event, error) { return nil, nil },
		upsertFn: func(in service.EventIn) (service.Event, error) {
			return service.Event{
				Source:    in.Source,
				Type:      in.Type,
				Count:     1,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		getByID: func(id string) (service.Event, error) { return service.Event{}, errors.New("nope") },
	}
	h := controllers.NewEventHandler(s)
	r := newRouter(h)

	body := bytes.NewBufferString(`{"source":"app.web","type":"user_login","payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/events", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"count":1`)) {
		t.Fatalf("expected count=1, got %s", w.Body.String())
	}
}
