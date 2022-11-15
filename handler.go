package main

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
)

type (
	ctxKeyHttpRequest   struct{}
	ctxKeyRequestHeader struct{}
)

type Handler[T, V any] func(ctx context.Context, t *T) (*V, error)

func Handle[T, V any](h Handler[T, V]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req T
		err := SetStructValue(r, &req)
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

func SetStructValue(r *http.Request, t any) error {
	fields := reflect.TypeOf(t).Elem()
	values := reflect.ValueOf(t)

	num := fields.NumField()
	for i := 0; i < num; i++ {

		field := fields.Field(i)
		value := values.Elem().Field(i)

		val, ok := GetFieldValueFromTag(r, field)
		if !ok || val == "" {
			continue
		}

		err := SetFieldValue(value, val)
		if err != nil {
			return err
		}
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		return json.NewDecoder(r.Body).Decode(t)
	default:
		return nil
	}
}

func SetFieldValue(value reflect.Value, val string) error {
	if !value.CanSet() {
		return nil
	}

	switch value.Kind() {
	case reflect.String:
		value.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		value.SetInt(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		value.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		value.SetBool(v)
	}
	return nil
}

func GetFieldValueFromTag(r *http.Request, field reflect.StructField) (string, bool) {
	var (
		tag string
		ok  bool
	)

	if tag, ok = field.Tag.Lookup("route_var"); ok {
		val := mux.Vars(r)[tag]
		return val, true
	}

	if tag, ok = field.Tag.Lookup("header"); ok {
		val := r.Header.Get(tag)
		return val, true
	}

	if tag, ok = field.Tag.Lookup("param"); ok {
		val := r.URL.Query().Get(tag)
		return val, true
	}

	return "", false
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
