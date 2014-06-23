package control

import (
	"labix.org/v2/mgo/bson"
	"time"
)

type User struct {
	Id       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Username string        `json:"username"`
	Password string        `json:"password"`
}

type Version struct {
	Version  string    `json:"version"`
	Filename string    `json:"filename"`
	Created  time.Time `json:"created"`
}

type Worker struct {
	Port      uint32 `json:"port"`
	Host      string `json:"host"`
	Container string `json:"container"`
	Version   string `json:"version"`
}

type App struct {
	Id          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Path        string        `json:"path"`
	Domains     string        `json:"domains"`
	Capacity    uint32        `json:"capacity"`
	Versions    []Version     `json:"versions"`
	Workers     []Worker      `json:"workers"`
}
