package usecase

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/yuriadams/lear/internal/model"
	"github.com/yuriadams/lear/internal/repository"
	"github.com/yuriadams/lear/internal/service"
)

type BookUsecase struct {
	Repo   *repository.BookRepository
	Logger *service.Logger
}

type BookAnalysis struct {
	Summary    string
	Sentiment  string
	Characters []string
}

func NewBookUsecase(repo *repository.BookRepository) *BookUsecase {
	return &BookUsecase{Repo: repo, Logger: service.NewLogger("[BookUsecase]")}
}

func (u *BookUsecase) FetchBook(gutenbergID int) (*model.Book, error) {
	u.Logger.SetTags(fmt.Sprintf("[book-%d]", gutenbergID))

	existingBook, err := u.Repo.GetBookByID(gutenbergID)
	if err != nil {
		u.Logger.LogError("Failed fetch Book", err)
		return nil, err
	}

	if existingBook != nil {
		u.Logger.LogInfo("Returning existing book from cashing")
		return existingBook, nil
	}

	metadata, err := service.ScrapeMetadata(gutenbergID)
	if err != nil {
		u.Logger.LogError("Failed to fetch metadata", err)
		return nil, err
	}

	u.Logger.LogInfo("Metadata fetched successfully")

	contentURL := fmt.Sprintf("https://www.gutenberg.org/cache/epub/%d/pg%d.txt", gutenbergID, gutenbergID)
	resp, err := http.Get(contentURL)
	if err != nil {
		u.Logger.LogError("Failed to fetch Content", err)
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	u.Logger.LogInfo("Content fetched successfully")

	book := &model.Book{
		GutenbergID: gutenbergID,
		Content:     string(content),
		Metadata:    *metadata,
	}

	err = u.Repo.SaveBook(book)
	if err != nil {
		u.Logger.LogError("Failed to save book", err)
		return nil, err
	}

	u.Logger.LogInfo("Book saved successfully and returning")
	return book, nil
}
