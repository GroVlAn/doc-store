package core

import "time"

type User struct {
	ID       string `json:"-" bson:"_id"`
	Login    string `json:"login" bson:"login"`
	Password string `json:"pswd" bson:"password"`
}

type AccessToken struct {
	ID       string    `bson:"_id"`
	Token    string    `bson:"token"`
	StartTTL time.Time `bson:"start_ttl"`
	EndTTl   time.Time `bson:"end_ttl"`
	UserID   string    `bson:"user_id"`
}
