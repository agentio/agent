//
//  Copyright 2014 Radtastical Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//
package agent

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
	Capacity    uint32        `json:"capacity"`
	Paths       []string      `json:"paths"`
	Domains     []string      `json:"domains"`
	Versions    []Version     `json:"versions"`
	Workers     []Worker      `json:"workers"`
}

type Connection struct {
        Service     string
        Credentials string
}

