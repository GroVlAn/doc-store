package core

import "time"

type DocumentRequest struct {
	Meta Document `json:"meta"`
	Json []byte   `json:"json"`
	File []byte   `json:"file"`
}

type Document struct {
	ID      string    `json:"-" bson:"_id"`
	Name    string    `json:"name" bson:"name"`
	Mine    string    `json:"mine" bson:"mine"`
	File    bool      `json:"file" bson:"file"`
	Token   string    `json:"token" bson:"-"`
	Json    []byte    `json:"-" bson:"json"`
	Public  bool      `json:"public" bson:"public"`
	Created time.Time `json:"created" bson:"created"`
	Grant   []string  `json:"grant" bson:"grant"`
}

type DocumentFilter struct {
	Token string `json:"token"`
	Login string `json:"login"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Limit int64  `json:"limit"`
}
