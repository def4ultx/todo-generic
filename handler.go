package main

import (
	"context"
	"encoding/json"
	"net/http"
)

type Binder[T any] interface {
	Bind(*http.Request) (T, error)
}

type NoOpBinder[T any] struct{}

func (x NoOpBinder[T]) Bind(r *http.Request) (T, error) {
	var t T
	return t, nil
}

type JsonBinder[T any] struct{}

func (x JsonBinder[T]) Bind(r *http.Request) (T, error) {
	var t T
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		return t, err
	}
	return t, nil
}

type (
	ctxKeyHttpRequest   struct{}
	ctxKeyRequestHeader struct{}
)

type Handler[T, V any] func(ctx context.Context, t *T) (*V, error)

func Handle[T Binder[T], V any](h Handler[T, V]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var binder T
		req, err := binder.Bind(r)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		validator, err := GetValidator(r.Context())
		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err)
			return
		}
		err = validator.Struct(req)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKeyRequestHeader{}, r.Header)
		ctx = context.WithValue(ctx, ctxKeyHttpRequest{}, r)
		resp, err := h(ctx, &req)
		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		WriteResponse(w, http.StatusOK, resp)
	}
}

func WriteErrorResponse(w http.ResponseWriter, code int, err error) {
	resp := struct {
		Msg string
	}{
		Msg: err.Error(),
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func WriteResponse(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
