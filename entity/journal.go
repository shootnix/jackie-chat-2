package entity

import (
	"github.com/shootnix/jackie-chat-2/io"
)

type Journal struct {
	ID         int64  `db:"id"`
	WorkerName string `db:"worker_name"`
	CTime      string `db:"ctime"`
	MessageID  int64  `db:"message_id"`
}

func NewJournal(worker_name string, message_id int64) *Journal {
	j := new(Journal)
	j.WorkerName = worker_name
	j.MessageID = message_id

	return j
}

func (j *Journal) Insert() error {
	sql := `

        INSERT INTO journal (worker_name, message_id) 
        VALUES ($1, $2)
        RETURNING id
    
    `

	row := io.GetPg().Conn.QueryRow(sql, j.WorkerName, j.MessageID)
	if err := row.Scan(&j.ID); err != nil {
		return err
	}

	return nil
}
