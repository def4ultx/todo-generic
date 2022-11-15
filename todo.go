package main

import (
	"context"
)

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GetRequest struct {
	ID int `route_var:"id"`
}

type GetResponse struct {
	Todo
}

type ListRequest struct{}

type ListResponse struct {
	Data []Todo `json:"data"`
}

type CreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateResponse struct {
	ID int `json:"id"`
}

type UpdateRequest struct {
	ID          int    `route_var:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateResponse struct{}

type DeleteRequest struct {
	ID int `route_var:"id"`
}

type DeleteResponse struct{}

func GetTodo(ctx context.Context, req *GetRequest) (*GetResponse, error) {

	db, err := GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var resp GetResponse
	err = db.QueryRow(ctx, "select id, title, description from todo where id = $1", req.ID).Scan(&resp.ID, &resp.Title, &resp.Description)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func ListTodo(ctx context.Context, req *ListRequest) (*ListResponse, error) {

	db, err := GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, "select id, title, description from todo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resp := ListResponse{
		Data: make([]Todo, 0),
	}
	for rows.Next() {
		var t Todo
		if err = rows.Scan(&t.ID, &t.Title, &t.Description); err != nil {
			return nil, err
		}
		resp.Data = append(resp.Data, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &resp, nil

}

func CreateTodo(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {

	db, err := GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	resp := CreateResponse{}
	err = db.QueryRow(ctx, `insert into todo(title, description) values ($1, $2) returning id`, req.Title, req.Description).Scan(&resp.ID)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}
func UpdateTodo(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {

	db, err := GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(ctx, `update todo set title = $2, description = $3 where id = $1`, req.ID, req.Title, req.Description)
	if err != nil {
		return nil, err
	}
	return &UpdateResponse{}, nil
}

func DeleteTodo(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {

	db, err := GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(ctx, `delete from todo where id = $1`, req.ID)
	if err != nil {
		return nil, err
	}
	return &DeleteResponse{}, nil
}
