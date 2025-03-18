package entities

type Body struct {
	Collection string `json:"Collection,omitempty"`
	Table      string `json:"Table,omitempty"`
	Method     string `json:"Method"`
	Query      any    `json:"Query"`
	Document   any    `json:"Document"`
	Options    any    `json:"Options"`
	Order      any    `json:"Order,omitempty"`
}
type ProcessData[T any] struct {
	Body    Body   `json:"Body"`
	RawData string `json:"RawData,omitempty"`
	Data    T      `json:"Data,omitempty"`
}
