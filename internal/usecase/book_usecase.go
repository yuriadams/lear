package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/yuriadams/lear/internal/domain"
	"github.com/yuriadams/lear/internal/repository"
	"github.com/yuriadams/lear/internal/service"
)

type IBookUsecase interface {
	FetchBook(gutenbergID int) (*domain.Book, error)
	FetchAllBooks() ([]domain.Book, error)
}

type BookUsecase struct {
	Repo    repository.IBookRepository
	Scraper service.IScraperMetadata
	Logger  *service.Logger
}

func NewBookUsecase(repo repository.IBookRepository, scraper service.IScraperMetadata) *BookUsecase {
	return &BookUsecase{Repo: repo, Scraper: scraper, Logger: service.NewLogger("[BookUsecase]")}
}

func (u *BookUsecase) FetchAllBooks() ([]domain.Book, error) {
	return u.Repo.GetAllBooks()
}

func (u *BookUsecase) FetchBook(gutenbergID int) (*domain.Book, error) {
	u.Logger.SetTags(fmt.Sprintf("[book-%d]", gutenbergID))

	existingBook, err := u.Repo.GetBookByID(gutenbergID)
	if err != nil {
		u.Logger.LogError("Failed to fetch book", err)
		return nil, err
	}

	if existingBook != nil {
		u.Logger.LogInfo("Returning existing book from cache")
		return existingBook, nil
	}

	var wg sync.WaitGroup
	contentCh := make(chan []byte, 1)
	metadataCh := make(chan []byte, 1)
	errCh := make(chan error, 2)

	wg.Add(2)

	go u.fetchBookContent(contentCh, errCh, gutenbergID, &wg)
	go u.fetchBookMetadata(metadataCh, errCh, gutenbergID, &wg)

	wg.Wait()

	close(contentCh)
	close(metadataCh)
	close(errCh)

	var errorMsg string
	for err := range errCh {
		if err != nil {
			errorMsg += err.Error() + "; "
		}
	}

	if errorMsg != "" {
		u.Logger.LogError("Failed to fetch book data", errors.New(errorMsg))
		return nil, errors.New(errorMsg)
	}

	content := <-contentCh
	metadataJSON := <-metadataCh

	var metadata domain.Metadata
	err = json.Unmarshal(metadataJSON, &metadata)
	if err != nil {
		u.Logger.LogError("Failed to decode metadata JSON", err)
		return nil, err
	}

	book := &domain.Book{
		GutenbergID: gutenbergID,
		Content:     string(content),
		Metadata:    metadata,
	}

	if err := u.Repo.SaveBook(book); err != nil {
		u.Logger.LogError("Failed to save book", err)
		return nil, err
	}

	u.Logger.LogInfo("Book saved successfully")
	return book, nil
}

func (u *BookUsecase) fetchBookContent(ch chan []byte, errCh chan error, gutenbergID int, wg *sync.WaitGroup) {
	defer wg.Done()

	contentURL := fmt.Sprintf("https://www.gutenberg.org/cache/epub/%d/pg%d.txt", gutenbergID, gutenbergID)
	resp, err := http.Get(contentURL)
	if err != nil {
		errCh <- fmt.Errorf("failed to fetch content: %w", err)
		return
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		errCh <- fmt.Errorf("failed to read content: %w", err)
		return
	}

	ch <- content
	u.Logger.LogInfo("Content fetched successfully")
}

func (u *BookUsecase) fetchBookMetadata(ch chan []byte, errCh chan error, gutenbergID int, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("https://www.gutenberg.org/ebooks/%d", gutenbergID)

	metadata, err := u.Scraper.ScrapeMetadata(url)
	if err != nil {
		errCh <- fmt.Errorf("failed to fetch metadata: %w", err)
		return
	}

	ch <- metadata
	u.Logger.LogInfo("Metadata fetched successfully")
}
