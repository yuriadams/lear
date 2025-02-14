package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yuriadams/lear/internal/domain"
	"github.com/yuriadams/lear/internal/usecase"
)

type MockBookRepository struct {
	mock.Mock
}

func (m *MockBookRepository) SaveBook(book *domain.Book) error {
	args := m.Called(book)
	return args.Error(0)
}

func (m *MockBookRepository) GetBookByID(gutenbergID int) (*domain.Book, error) {
	args := m.Called(gutenbergID)
	book, _ := args.Get(0).(*domain.Book)
	return book, args.Error(1)
}

func (m *MockBookRepository) GetAllBooks() ([]domain.Book, error) {
	args := m.Called()
	return args.Get(0).([]domain.Book), args.Error(1)
}

type MockScraperMetadata struct {
	mock.Mock
}

func (m *MockScraperMetadata) ScrapeMetadata(url string) (*domain.Metadata, error) {
	args := m.Called(url)
	metadata, _ := args.Get(0).(*domain.Metadata)
	return metadata, args.Error(1)
}

func TestFetchBook(t *testing.T) {
	mockRepo := &MockBookRepository{}
	mockRepo.On("GetBookByID", 12345).Return(nil, nil)
	mockRepo.On("SaveBook", mock.Anything).Return(nil)

	mockScraper := &MockScraperMetadata{}
	mockScraper.On("ScrapeMetadata", "https://www.gutenberg.org/ebooks/12345").Return(&domain.Metadata{
		Title:    "Test Book Title",
		Author:   "Test Author",
		Language: "English",
		Subject:  "Fiction",
		Credits:  "Test Credits",
		Summary:  "This is a test summary.",
	}, nil)

	bookUsecase := usecase.NewBookUsecase(mockRepo, mockScraper)

	book, err := bookUsecase.FetchBook(12345)

	assert.NoError(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, 12345, book.GutenbergID)
	assert.Equal(t, "Test Book Title", book.Metadata.Title)
	assert.Equal(t, "Test Author", book.Metadata.Author)
	assert.Equal(t, "English", book.Metadata.Language)
	assert.Equal(t, "Fiction", book.Metadata.Subject)
	assert.Equal(t, "This is a test summary.", book.Metadata.Summary)

	mockRepo.AssertExpectations(t)
	mockScraper.AssertExpectations(t)
}
