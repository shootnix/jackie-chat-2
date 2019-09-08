package entity

import (
	"errors"
	"github.com/shootnix/jackie-chat-2/io"
	"github.com/shootnix/jackie-chat-2/logger"
	"golang.org/x/crypto/bcrypt"
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

        SELECT id, name, default_bot, password
          FROM users u 
         WHERE is_deleted = false
           AND is_active = true
           AND name = $1

    `
	log := logger.GetLogger()

	hashPwd, _ := hashPassword("12345")
	log.Debug("password = " + hashPwd)

	row := io.GetPg().Conn.QueryRow(sql, name)
	if err := row.Scan(&u.ID, &u.Name, &u.DefaultBot, &u.Password); err != nil {
		return u, err
	}

	if chechPassword(u.Password, password) == false {
		return u, errors.New("Wrong Credentials")
	}

	return u, nil
}

func hashPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 14)
	return string(bytes), err
}

func chechPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
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
