package entity

import (
	"github.com/shootnix/jackie-chat-2/io"
)

type User struct {
	ID         int64  `db:"id"`
	Name       string `db:"name"`
	Password   string `db:"name"`
	DefaultBot int64  `db:"default_bot"`
	IsActive   bool   `db:"is_active"`
	IsDeleted  bool   `db:"is_deleted"`
}

func LoginUser(name, password string) (*User, error) {
	u := new(User)
	sql := `

        SELECT id, name, default_bot
          FROM users u 
         WHERE is_deleted = false
           AND is_active = true
           AND name = $1
           AND password = $2

    `

	row := io.GetPg().Conn.QueryRow(sql, name, password)
	if err := row.Scan(&u.ID, &u.Name, &u.DefaultBot); err != nil {
		return u, err
	}

	return u, nil
}

func FindUser(name string) (*User, error) {
	u := new(User)
	sql := `

        SELECT id, name, default_bot
          FROM users u 
         WHERE is_deleted = false
           AND is_active = true
           AND name = $1

    `

	row := io.GetPg().Conn.QueryRow(sql, name)
	if err := row.Scan(&u.ID, &u.Name, &u.DefaultBot); err != nil {
		return u, err
	}

	return u, nil
}
