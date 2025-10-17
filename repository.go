package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Task — модель для сканирования результатов SELECT
type Task struct {
	ID        int
	Title     string
	Done      bool
	CreatedAt time.Time
}

type Repo struct{ DB *sql.DB }

func NewRepo(db *sql.DB) *Repo { return &Repo{DB: db} }

// CreateTask — параметризованный INSERT с возвратом id
func (r *Repo) CreateTask(ctx context.Context, title string) (int, error) {
	const q = `INSERT INTO tasks (title) VALUES ($1) RETURNING id;`
	var id int
	if err := r.DB.QueryRowContext(ctx, q, title).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// ListTasks — базовый SELECT всех задач
func (r *Repo) ListTasks(ctx context.Context) ([]Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks ORDER BY id;`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil { return nil, err }
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil { return nil, err }
		out = append(out, t)
	}
	return out, rows.Err()
}

// --- Доп. задания ---

// ListDone — возвращает выполненные (done=true) или невыполненные (done=false) задачи.
func (r *Repo) ListDone(ctx context.Context, done bool) ([]Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks WHERE done = $1 ORDER BY id;`
	rows, err := r.DB.QueryContext(ctx, q, done)
	if err != nil { return nil, err }
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil { return nil, err }
		out = append(out, t)
	}
	return out, rows.Err()
}

// FindByID — возвращает задачу по id
func (r *Repo) FindByID(ctx context.Context, id int) (*Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks WHERE id = $1;`
	var t Task
	err := r.DB.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // не нашли
		}
		return nil, err
	}
	return &t, nil
}

// CreateMany — массовая вставка через транзакцию
func (r *Repo) CreateMany(ctx context.Context, titles []string) error {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil { return err }
	defer func() {
		if err != nil { _ = tx.Rollback() }
	}()

	const q = `INSERT INTO tasks (title) VALUES ($1);`
	for _, title := range titles {
		if _, err = tx.ExecContext(ctx, q, title); err != nil {
			return err
		}
	}
	return tx.Commit()
}
