package model

type PaginateInput struct {
	Page  int `query:"page" example:"1" doc:"Page number" required:"true" default:"0" minimum:"0"`
	Limit int `query:"limit" example:"10" doc:"Page size" required:"true" default:"10" minimum:"1"`
}
