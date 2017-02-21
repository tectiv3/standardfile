package models

import (
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/deckarep/golang-set"
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
	Items       Items  `json:"items"`
	SyncToken   string `json:"sync_token"`
	CursorToken string `json:"cursor_token"`
	Limit       int    `json:"limit"`
}

type unsaved struct {
	Item
	error
}

//SyncResponse - type for response
type SyncResponse struct {
	Retrieved   Items     `json:"retrieved_items"`
	Saved       Items     `json:"saved_items"`
	Unsaved     []unsaved `json:"unsaved_items"`
	SyncToken   string    `json:"sync_token"`
	CursorToken string    `json:"cursor_token,omitempty"`
}

const minConflictInterval = 5.0

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
func (i *Item) save() error {
	if i.Uuid == "" || !i.Exists() {
		return i.create()
	}
	return i.update()
}

func (i *Item) create() error {
	if i.Uuid == "" {
		i.Uuid = uuid.NewV4().String()
	}
	i.Created_at = time.Now()
	i.Updated_at = time.Now()
	i.User_uuid = Auth.User.Uuid
	log.Println("Create:", i)
	return db.Query("INSERT INTO `items` (`uuid`, `user_uuid`, content,  content_type, enc_item_key, auth_hash, deleted, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,?)", i.Uuid, i.User_uuid, i.Content, i.Content_type, i.Enc_item_key, i.Auth_hash, i.Deleted, i.Created_at, i.Updated_at)
}

func (i Item) copy() Item {
	i.Uuid = uuid.NewV4().String()
	i.Updated_at = time.Now()
	return i
}

func (i *Item) delete() error {
	if i.Uuid == "" {
		return fmt.Errorf("Trying to delete unexisting item")
	} else {
		// find_or_create by uuid
	}
	return nil
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
func (user User) SyncItems(request SyncRequest) (SyncResponse, error) {

	response := SyncResponse{
		Retrieved:   Items{},
		Saved:       Items{},
		Unsaved:     []unsaved{},
		SyncToken:   GetTokenFromTime(time.Now()),
		CursorToken: "",
	}

	if request.Limit == 0 {
		request.Limit = 100000
	}

	response.SyncToken = GetTokenFromTime(time.Now())
	var err error
	var items Items
	if request.CursorToken != "" {
		items, err = user.getItemsFromDate(GetTimeFromToken(request.CursorToken))
	} else if request.SyncToken != "" {
		items, err = user.getItemsOlder(GetTimeFromToken(request.SyncToken))
	} else {
		items, err = user.getItems(request.Limit)
		response.CursorToken = GetTokenFromTime(items[len(items)-1].Updated_at)
	}
	log.Println(items)
	if err != nil {
		return response, err
	}
	response.Saved, response.Unsaved, err = request.Items.save()
	if err != nil {
		return response, err
	}
	if len(response.Saved) > 0 {
		response.SyncToken = GetTokenFromTime(response.Saved[0].Updated_at)
	}
	// manage conflicts
	saved := mapset.NewSet()
	retreived := mapset.NewSet()
	for _, item := range response.Retrieved {
		retreived.Add(item.Uuid)
	}
	for _, item := range response.Saved {
		saved.Add(item.Uuid)
	}
	conflicts := saved.Intersect(retreived)
	log.Println("Conflicts", conflicts)
	// saved items take precedence, retrieved items are duplicated with a new uuid
	for _, uuid := range conflicts.ToSlice() {
		// if changes are greater than minConflictInterval seconds apart, create conflicted copy, otherwise discard conflicted
		log.Println(uuid)
		savedCopy := response.Saved.find(uuid.(string))
		retreivedCopy := response.Retrieved.find(uuid.(string))

		if savedCopy.isConflictedWith(retreivedCopy) {
			log.Printf("Creating conflicted copy of %v\n", uuid)
			dupe := retreivedCopy.copy()
			response.Retrieved = append(response.Retrieved, dupe)
		}
		response.Retrieved.delete(uuid.(string))
	}

	return response, nil
}

func (i Item) isConflictedWith(copy Item) bool {
	diff := math.Abs(float64(i.Updated_at.Unix() - copy.Updated_at.Unix()))
	return diff > minConflictInterval
}

func (items Items) save() (Items, []unsaved, error) {
	var savedItems Items
	var unsavedItems []unsaved

	if len(items) == 0 {
		return savedItems, unsavedItems, nil
	}

	for _, item := range items {
		log.Println(item)
		var err error
		if item.Deleted {
			err = item.delete()
		} else {
			err = item.save()
		}
		if err != nil {
			unsavedItems = append(unsavedItems, unsaved{item, err})
		} else {
			savedItems = append(savedItems, item)
		}
	}
	return savedItems, unsavedItems, nil
}

func (items Items) find(uuid string) Item {
	for _, item := range items {
		if item.Uuid == uuid {
			return item
		}
	}
	return Item{}
}

func (items *Items) delete(uuid string) {
	position := 0
	for i, item := range *items {
		if item.Uuid == uuid {
			position = i
			break
		}
	}
	(*items) = (*items)[:position:position]
}
