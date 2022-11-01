package mysql

import (
	"database/sql"
	"errors"

	"github.com/kientink26/golang-started/pkg/models"
)

type ThreadModel struct {
	DB *sql.DB
}

// This will insert a new thread into the database.
func (m *ThreadModel) Insert(topic string, userId int) (int, error) {
	stmt := `INSERT INTO threads (topic, user_id, created)
			VALUES(?, ?, UTC_TIMESTAMP())`

	result, err := m.DB.Exec(stmt, topic, userId)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// This will return a specific thread based on its id.
func (m *ThreadModel) Get(id int) (*models.Thread, error) {
	stmt := `SELECT t.id, t.topic, t.created, u.id, u.name
			FROM threads as t
			INNER JOIN users as u
			ON t.user_id = u.id
			WHERE t.id = ?`

	row := m.DB.QueryRow(stmt, id)

	thread := &models.Thread{User: &models.User{}} // avoid nil pointer dereference
	err := row.Scan(&thread.ID, &thread.Topic, &thread.Created, &thread.User.ID, &thread.User.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything went OK then return the Thread object.
	return thread, nil

}

// This will return the 10 most recently created threads.
func (m *ThreadModel) Latest() ([]*models.Thread, error) {
	stmt := `SELECT t.id, t.topic, t.created, u.id, u.name
			FROM threads as t
			INNER JOIN users as u
			ON t.user_id = u.id
			ORDER BY t.created DESC
			LIMIT 10`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	threads := []*models.Thread{}
	for rows.Next() {
		s := &models.Thread{User: &models.User{}}
		err = rows.Scan(&s.ID, &s.Topic, &s.Created, &s.User.ID, &s.User.Name)
		if err != nil {
			return nil, err
		}
		threads = append(threads, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return threads, nil
}
