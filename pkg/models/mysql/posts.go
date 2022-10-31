package mysql

import (
	"database/sql"
	"errors"

	"github.com/kientink26/golang-started/pkg/models"
)

type PostModel struct {
	DB *sql.DB
}

func (m *PostModel) Insert(body string, userId int, threadId int) (int, error) {
	stmt := `INSERT INTO posts (body, user_id, thread_id, created)
			VALUES(?, ?, ?, UTC_TIMESTAMP())`

	result, err := m.DB.Exec(stmt, body, userId, threadId)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *PostModel) Get(id int) (*models.Post, error) {
	stmt := `SELECT id, body, user_id, created FROM posts WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	post := &models.Post{User: &models.User{}} // avoid nil pointer dereference
	err := row.Scan(&post.ID, &post.Body, &post.User.ID, &post.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return post, nil

}

func (m *PostModel) Delete(id int, userId int) error {
	stmt := `DELETE from posts WHERE id = ? AND user_id = ?`

	_, err := m.DB.Exec(stmt, id, userId)
	return err
}

func (m *PostModel) Latest(threadId int) ([]*models.Post, error) {
	stmt := `SELECT id, body, user_id, created FROM posts WHERE thread_id = ? ORDER BY created DESC`
	rows, err := m.DB.Query(stmt, threadId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.Post{}
	for rows.Next() {
		s := &models.Post{User: &models.User{}} // avoid nil pointer dereference
		err = rows.Scan(&s.ID, &s.Body, &s.User.ID, &s.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}
