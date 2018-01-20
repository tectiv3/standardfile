package main

import (
	"database/sql"
	m "github.com/remind101/migrate"
	"github.com/tectiv3/standardfile/db"
	"log"
	"time"
)

//Migrate performs migration
func Migrate(dbpath string) {
	db.Init(dbpath)
	migrations := getMigrations()
	err := m.Exec(db.DB(), m.Up, migrations...)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Done")
}

func getMigrations() []m.Migration {
	migrations := []m.Migration{
		{
			ID: 1,
			Up: m.Queries([]string{
				"ALTER TABLE users ADD COLUMN pw_auth varchar(255);",
				"ALTER TABLE users ADD COLUMN pw_salt varchar(255);",
			}),
			Down: func(tx *sql.Tx) error {
				// It's not possible to remove a column with sqlite.
				return nil
			},
		},
		{
			ID: 2,
			Up: func(tx *sql.Tx) error {
				users := []User{}
				db.Select("SELECT * FROM `users`", &users)
				log.Println("Got", len(users), "users to update")
				for _, u := range users {
					if u.Email == "" || u.PwNonce == "" {
						continue
					}
					if _, err := tx.Exec("UPDATE `users` SET `pw_salt`=?, `updated_at`=? WHERE `uuid`=?", getSalt(u.Email, u.PwNonce), time.Now(), u.UUID); err != nil {
						log.Println(err)
					}
				}

				return nil
			},
			Down: func(tx *sql.Tx) error {
				return nil
			},
		},
	}
	return migrations
}
