package repository

import (
	"database/sql"
	"errors"

	"github.com/yuriadams/lear/internal/domain"
)

type IBookRepository interface {
	GetBookByID(gutenbergID int) (*domain.Book, error)
	SaveBook(book *domain.Book) error
}

type BookRepository struct {
	DB *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{DB: db}
}

func (r *BookRepository) GetBookByID(gutenbergID int) (*domain.Book, error) {
	var book domain.Book
	query := `SELECT id, gutenberg_id, content, metadata FROM books WHERE gutenberg_id = $1`
	err := r.DB.QueryRow(query, gutenbergID).Scan(&book.ID, &book.GutenbergID, &book.Content, &book.Metadata)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &book, nil
}

func (r *BookRepository) SaveBook(book *domain.Book) error {
	query := `INSERT INTO books (gutenberg_id, content, metadata) VALUES ($1, $2, $3) RETURNING id`
	return r.DB.QueryRow(
		query,
		book.GutenbergID,
		book.Content,
		book.Metadata,
	).Scan(&book.ID)
}
