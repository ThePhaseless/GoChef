package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	custom_middlewares "github.com/ThePhaseless/GoChef/middlewares"
	"github.com/ThePhaseless/GoChef/model"
	"github.com/ThePhaseless/GoChef/query"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

type InBody[T any] struct {
	Body T `json:"body"`
}

func WrapInBody[T any](body T) *InBody[T] {
	return &InBody[T]{
		Body: body,
	}
}

func main() {
	// Create a new router & API
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("My API", "1.0.0"))

	sqliteDb := sqlite.Open("test.db")
	db, err := gorm.Open(sqliteDb, &gorm.Config{})
	db.Raw("PRAGMA journal_mode=WAL;")
	_ = db.AutoMigrate(&model.User{})

	if err != nil {
		panic(err)
	}

	GenerateQuery(db)

	query.SetDefault(db)

	// Register GET /greeting/{name}
	huma.Register(api, huma.Operation{
		OperationID: "get-greeting",
		Method:      http.MethodGet,
		Path:        "/greeting/{name}",
		Summary:     "Get a greeting",
		Description: "Get a greeting for a person by name.",
		Tags:        []string{"Greetings"},
		Middlewares: huma.Middlewares{custom_middlewares.RateLimitMiddleware(api)},
	}, func(ctx context.Context, input *struct {
		Name string `path:"name" maxLength:"30" example:"world" doc:"Name to greet"`
	}) (*model.GreetingOutput, error) {
		resp := &model.GreetingOutput{}
		resp.Body.Message = fmt.Sprintf("Hello, %s!", input.Name)

		err := query.User.Create(&model.User{Name: input.Name})

		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}

		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-names",
		Method:      http.MethodGet,
		Path:        "/names",
		Summary:     "Get names",
		Description: "Get names.",
		Tags:        []string{"Names"},
	}, func(ctx context.Context, i *model.PaginateInput) (*InBody[[]*model.User], error) {
		users, _, err := query.User.FindByPage((i.Page)*i.Limit, i.Limit)

		if err != nil {
			return nil, huma.Error400BadRequest(err.Error())
		}

		return WrapInBody(users), nil
	})

	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-name",
			Method:      http.MethodGet,
			Path:        "/names/{id}",
			Summary:     "Get name",
			Description: "Get name.",
			Tags:        []string{"Names"},
		},
		func(ctx context.Context, i *struct {
			ID string `path:"id"`
		}) (*model.User, error) {
			return nil, huma.Error406NotAcceptable("You know why...")
		})

	huma.Register(
		api,
		huma.Operation{
			OperationID: "upload-name",
			Method:      http.MethodPost,
			Path:        "/upload",
			Summary:     "Upload name",
			Description: "Upload a name in file to greet",
			Tags:        []string{"Names"},
		},
		func(ctx context.Context, input *struct {
			RawBody multipart.Form
		}) (*InBody[string], error) {
			file, _ := (input.RawBody.File["filename"][0]).Open()

			var buf bytes.Buffer
			_, err = io.Copy(&buf, file)
			if err != nil {
				return nil, huma.Error500InternalServerError(err.Error())
			}

			// The file content is in buf.Bytes()
			content := buf.String()

			err = query.User.Create(&model.User{Name: content})

			if err != nil {
				return nil, huma.Error500InternalServerError(err.Error())
			}

			return WrapInBody("Hello " + content), nil
		},
	)

	// Start the server!
	err = http.ListenAndServe("127.0.0.1:8888", router)
	if err != nil {
		log.Fatal(err)
	}
}
