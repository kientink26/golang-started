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
	stmt := `SELECT id, topic, user_id, created FROM threads WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	thread := &models.Thread{User: &models.User{}} // avoid nil pointer dereference
	err := row.Scan(&thread.ID, &thread.Topic, &thread.User.ID, &thread.Created)
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
	stmt := `SELECT id, topic, user_id, created FROM threads ORDER BY created DESC LIMIT 10`
	// Use the Query() method on the connection pool to execute our
	// SQL statement. This returns a sql.Rows resultset containing the result of
	// our query.
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Initialize an empty slice to hold the models.Thread objects.
	threads := []*models.Thread{}
	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by the
	// rows.Scan() method.
	for rows.Next() {
		s := &models.Thread{User: &models.User{}} // avoid nil pointer dereference
		err = rows.Scan(&s.ID, &s.Topic, &s.User.ID, &s.Created)
		if err != nil {
			return nil, err
		}
		threads = append(threads, s)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the Threads slice.
	return threads, nil
}
