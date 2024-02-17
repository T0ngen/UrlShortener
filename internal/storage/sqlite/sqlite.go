package sqlite

import (
	
	"database/sql"
	
	"errors"
	"fmt"

	"net/http"

	"github.com/mattn/go-sqlite3"

	"url-shortener/internal/hashedApi"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		api TEXT,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt2, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY,
		username TEXT,
		password TEXT);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt2.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt3, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS tokens(
		id INTEGER PRIMARY KEY,
		token TEXT);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt3.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}


	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(api string,urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(api, url, alias) VALUES(?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(api, urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}



func (s *Storage) AddNewAPIToDb(api string) error{
	const op = "storage.sqlite.AddNewAPIToDb"

	stmt, err := s.db.Prepare("INSERT INTO tokens (token) Values (?) ")
	if err != nil{
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(api)

	if err != nil{
		return fmt.Errorf("%s: %w", op, err)
	}
	
	return nil
	
}

func (s *Storage) DeleteUrl(hashedapi string, alias string) (bool, error) {


	query := "DELETE FROM url WHERE api = ? AND alias = ?"

// Используйте метод Exec для выполнения запроса DELETE
	result, err := s.db.Exec(query, hashedapi, alias)
	if err != nil {
		return false, fmt.Errorf("error with delete: %v", err)
	}

	// Проверяем количество затронутых строк после выполнения запроса
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Если ни одна строка не была удалена, вернуть false и ошибку
		return false, fmt.Errorf("no rows deleted")
	}

	return true, nil

} 


func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resURL, nil
}


//TODO перенести в отдельную папку
func BasicAuthFromDB(s *Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
	  fn := func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()
  
		// Проверка, что username и password корректные
		query := `SELECT COUNT(1) FROM users WHERE username=? AND password=?`
		var count int
		err := s.db.QueryRow(query, username, password).Scan(&count)
		if err != nil || count == 0 {
		  http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		  return
		}
  
		// Прошли аутентификацию, вызываем следующий обработчик
		next.ServeHTTP(w, r)
	  }
	  return http.HandlerFunc(fn)
	}
  }



  




func APITokenAuthFromDB(s *Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
	  fn := func(w http.ResponseWriter, r *http.Request) {
		
		token := r.Header.Get("Authorization")
		hashedToken := hashedapi.HashApi(token)
		// Проверка, что токен существует и корректный
		query := `SELECT COUNT(1) FROM tokens WHERE token=?`
		var count int
		err := s.db.QueryRow(query, hashedToken).Scan(&count)
		if err != nil || count == 0 {
		  http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		  return
		}
  
		// Прошли аутентификацию, вызываем следующий обработчик
		next.ServeHTTP(w, r)
	  }
	  return http.HandlerFunc(fn)
	}
  }


  
// TODO: implement method
// func (s *Storage) DeleteURL(alias string) error