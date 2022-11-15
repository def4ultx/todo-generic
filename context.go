package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ValidatorMiddleware() func(h http.Handler) http.Handler {
	validate := validator.New()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(NewValidatorContext(r.Context(), validate))
			next.ServeHTTP(w, r)
		})
	}
}

type ctxKeyValidator struct{}

func NewValidatorContext(ctx context.Context, v *validator.Validate) context.Context {
	ctx = context.WithValue(ctx, ctxKeyValidator{}, v)
	return ctx
}

func GetValidator(ctx context.Context) (*validator.Validate, error) {
	db, ok := ctx.Value(ctxKeyValidator{}).(*validator.Validate)
	if !ok {
		return nil, errors.New("error: validator not found in ctx")
	}
	return db, nil
}

type ctxKeyDB struct{}

func NewDatabaseContext(ctx context.Context, db *pgxpool.Pool) context.Context {
	ctx = context.WithValue(ctx, ctxKeyDB{}, db)
	return ctx
}

func GetDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	db, ok := ctx.Value(ctxKeyDB{}).(*pgxpool.Pool)
	if !ok {
		return nil, errors.New("error: database not found in ctx")
	}
	return db, nil
}

func DatabaseMiddleware(db *pgxpool.Pool) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(NewDatabaseContext(r.Context(), db))
			next.ServeHTTP(w, r)
		})
	}
}
