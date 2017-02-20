package models

import (
	"time"

	"github.com/satori/go.uuid"
)

// Item - is an item type
type Item struct {
	Uuid         string    `json:"uuid"`
	User_uuid    string    `json:"user_uuid"`
	Content      string    `json:"content"`
	Enc_item_key string    `json:"enc_item_key"`
	Auth_hash    string    `json:"auth_hash"`
	Content_type string    `json:"content_type"`
	Deleted      bool      `json:"deleted"`
	Created_at   time.Time `json:"created_at"`
	Updated_at   time.Time `json:"updated_at"`
}

//Items - is an items slice
type Items []Item

//Save - save current item into DB
func (i *Item) Save() {
	if i.Uuid == "" {
		i.Uuid = uuid.NewV4().String()
		i.Created_at = time.Now()
	}
	i.Updated_at = time.Now()
}
