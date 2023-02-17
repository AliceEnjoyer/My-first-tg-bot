package sqlite

import (
	"context"
	"database/sql"
	"firstTGBot/lib/e"
	"firstTGBot/storage"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db      *sql.DB
	context context.Context
}

func New(path string) (*Storage, error) {
	DB, err := sql.Open("sqlite3", path) // открывает файл с базой данных указанной версии (sqlite3)
	if err != nil {
		return nil, e.Wrap("can not open DB", err)
	}
	if err := DB.Ping(); err != nil { // подключилась ли база данных корректно
		return nil, e.Wrap("can not connect to DB", err)
	}
	return &Storage{db: DB}, nil
}

func (s *Storage) Save(p *storage.Page) error {
	q := `INSERT INTO pages (url, user_name) VALUES(?, ?)`

	// если в двух сдовах, то контекст делает так, что бы таймаут был конкретно нами заданый
	if _, err := s.db.ExecContext(s.context, q, p.URL, p.UserName); err != nil { // VALUES(?, ?) >>> VALUES(p.URL, p.UserName)
		return e.Wrap("Can not save page", err)
	}

	return nil
}

func (s *Storage) PickRandom(userName string) (*storage.Page, error) {
	q := `SELECT url FROM pages WHERE user_name = ? ORDER BY RANDOM() LIMIT 1`

	var resUrl string // переменная, куда будет сохрантся результат

	// QueryRowContext возвращает структуру с данными из талицы. Так как они могут быть гигантстких размеров, будет не очень
	// удобно сразу поместить их в память. По этому нужно использовать метод Scan
	err := s.db.QueryRowContext(s.context, q, userName).Scan(&resUrl)

	if err == sql.ErrNoRows { // если пользователь ничего не сохранял
		return nil, storage.ErrNoSavedPages
	}
	if err != nil {
		return nil, e.Wrap("Can not pick page", err)
	}

	return &storage.Page{
		URL:      resUrl,
		UserName: userName,
	}, nil
}

func (s *Storage) Remove(page *storage.Page) error {
	q := `DELETE FROM pages WHERE url = ? AND user_name = ?`
	if _, err := s.db.ExecContext(s.context, q, page.URL, page.UserName); err != nil {
		return e.Wrap("Can not remove page", err)
	}
	return nil
}

// комментарии в стиле go doc!!!!

// IsExists cheks if page exists in storage. <- точка объязательно (go doc комманда)
func (s *Storage) IsExists(page *storage.Page) (bool, error) {
	q := `SELECT COUNT(*) FROM pages WHERE url = ? AND user_name = ?`
	var res int
	err := s.db.QueryRowContext(s.context, q, page.URL, page.UserName).Scan(&res)
	if err != nil {
		return false, e.Wrap("Can not check if page exists", err)
	}
	return res > 0, nil
}

// создает базу данных
func (s *Storage) Init(cntxt context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS pages (url TEXT, user_name TEXT)`
	if _, err := s.db.ExecContext(cntxt, q); err != nil {
		return e.Wrap("Can not create DB", err)
	}
	s.context = cntxt

	return nil
}
