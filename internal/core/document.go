package core

import "time"

type Document struct {
	ID      string    `json:"-" bson:"_id"`
	Name    string    `json:"name" bson:"name"`
	Mine    string    `json:"mine" bson:"mine"`
	File    bool      `json:"file" bson:"file"`
	Public  bool      `json:"public" bson:"public"`
	Created time.Time `json:"created" bson:"created"`
	Grant   []string  `json:"grant" bson:"grant"`
}
