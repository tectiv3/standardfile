package models

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
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

//SyncRequest - type for incoming sync request
type SyncRequest struct {
	Items       []Item `json:"items"`
	SyncToken   string `json:"sync_token"`
	CursorToken string `json:"cursor_token"`
	Limit       int    `json:"limit"`
}

//SyncResponse - type for response
type SyncResponse struct {
	Retrieved Items  `json:"retrieved_items"`
	Saved     Items  `json:"saved_items"`
	Unsaved   Items  `json:"unsaved_items"`
	SyncToken string `json:"sync_token"`
}

//LoadValue - hydrate struct from map
func (r *SyncRequest) LoadValue(name string, value []string) {
	switch name {
	case "items":
		r.Items = Items{}
	case "sync_token":
		r.SyncToken = value[0]
	case "cursor_token":
		r.CursorToken = value[0]
	case "limit":
		r.Limit, _ = strconv.Atoi(value[0])
	}
}

//LoadValue - hydrate struct from map
func (i *Item) LoadValue(name string, value []string) {
	switch name {
	case "uuid":
		i.Uuid = value[0]
	case "user_uuid":
		i.User_uuid = value[0]
	case "content":
		i.Content = value[0]
	case "enc_item_key":
		i.Enc_item_key = value[0]
	case "content_type":
		i.Content_type = value[0]
	case "auth_hash":
		i.Content_type = value[0]
	case "deleted":
		i.Deleted = (value[0] == "true")
	}
}

//Save - save current item into DB
func (i *Item) save() {
	if i.Uuid == "" {
		i.Uuid = uuid.NewV4().String()
		i.Created_at = time.Now()
	}
	i.Updated_at = time.Now()
}

//GetTokenFromTime - generates sync token for current time
func GetTokenFromTime(date time.Time) string {
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("1:%d", date.Unix())))
}

//GetTimeFromToken - retreive datetime from sync token
func GetTimeFromToken(token string) time.Time {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		log.Println(err)
		return time.Now()
	}
	log.Println(decoded)
	parts := strings.Split(string(decoded), ":")
	str, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Println(err)
		return time.Now()
	}
	return time.Time(time.Unix(int64(str), 0))
}

//SyncItems - sync manager
func SyncItems(input SyncRequest) (SyncResponse, error) {
	return SyncResponse{
		Items{},
		Items{},
		Items{},
		GetTokenFromTime(time.Now()),
	}, nil
}
