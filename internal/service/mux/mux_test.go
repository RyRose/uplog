package mux

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestErrorResponseWriter_WriteHeader_Success(t *testing.T) {
	w := httptest.NewRecorder()
	ew := &errorResponseWriter{
		ResponseWriter: w,
		ctx:            context.Background(),
		statusCode:     0,
	}

	ew.WriteHeader(http.StatusOK)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestErrorResponseWriter_WriteHeader_Error(t *testing.T) {
	w := httptest.NewRecorder()
	ew := &errorResponseWriter{
		ResponseWriter: w,
		ctx:            context.Background(),
		statusCode:     0,
	}

	ew.WriteHeader(http.StatusBadRequest)

	// Should not write header yet for error codes
	if ew.statusCode != http.StatusBadRequest {
		t.Errorf("expected statusCode %d, got %d", http.StatusBadRequest, ew.statusCode)
	}
}

func TestErrorResponseWriter_Write_Success(t *testing.T) {
	w := httptest.NewRecorder()
	ew := &errorResponseWriter{
		ResponseWriter: w,
		ctx:            context.Background(),
		statusCode:     http.StatusOK,
	}

	data := []byte("test data")
	n, err := ew.Write(data)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != len(data) {
		t.Errorf("expected to write %d bytes, wrote %d", len(data), n)
	}

	if w.Body.String() != "test data" {
		t.Errorf("expected body %q, got %q", "test data", w.Body.String())
	}
}

func TestWeb_Handle(t *testing.T) {
	mux := &http.ServeMux{}
	web := &Web{Mux: &Trace{Mux: mux}}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("web handler"))
	})

	web.Handle("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if body != "web handler" {
		t.Errorf("expected body %q, got %q", "web handler", body)
	}
}

func TestWeb_HandleFunc(t *testing.T) {
	mux := &http.ServeMux{}
	web := &Web{Mux: &Trace{Mux: mux}}

	web.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	body := w.Body.String()
	if body != "created" {
		t.Errorf("expected body %q, got %q", "created", body)
	}
}

func TestWeb_HandleError(t *testing.T) {
	mux := &http.ServeMux{}
	web := &Web{Mux: &Trace{Mux: mux}}

	web.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error message"))
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Should transform to UnprocessableEntity
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "400") {
		t.Errorf("expected body to contain status code, got %q", body)
	}
}

func TestErrorResponseWriter_MultipleWrites(t *testing.T) {
	w := httptest.NewRecorder()
	ew := &errorResponseWriter{
		ResponseWriter: w,
		ctx:            context.Background(),
		statusCode:     http.StatusOK,
	}

	ew.Write([]byte("first "))
	ew.Write([]byte("second"))

	body := w.Body.String()
	if body != "first second" {
		t.Errorf("expected body %q, got %q", "first second", body)
	}
}
