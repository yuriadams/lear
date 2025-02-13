package delivery_test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yuriadams/lear/internal/delivery"
	"github.com/yuriadams/lear/internal/domain"
)

// Mock Usecase and Service
type MockBookUsecase struct {
	mock.Mock
}

func (m *MockBookUsecase) FetchBook(id int) (*domain.Book, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Book), args.Error(1)
}

type MockAnalysisService struct {
	mock.Mock
}

func (m *MockAnalysisService) StreamTextAnalysis(w http.ResponseWriter, r *http.Request, content string) error {
	args := m.Called(w, r, content)
	return args.Error(0)
}

func createTestTemplates() *template.Template {
	tmpl, err := template.New("layout.html").Parse(`
		<!DOCTYPE html>
		<html>
		<head><title>{{.Title}}</title></head>
		<body>{{.Body}}</body>
		</html>
	`)
	if err != nil {
		panic(err)
	}

	_, err = tmpl.New("index.html").Parse("<h1>Welcome to Project King Lear Explorer</h1>")
	if err != nil {
		panic(err)
	}

	_, err = tmpl.New("show.html").Parse(`
		<h1>{{.Title}}</h1>
		<p>By {{.Author}}</p>
		<div>{{.Content}}</div>
	`)
	if err != nil {
		panic(err)
	}

	return tmpl
}

func TestBookHandler_Show(t *testing.T) {
	// Create mock dependencies
	mockUsecase := new(MockBookUsecase)
	mockService := new(MockAnalysisService)
	templates := createTestTemplates()

	handler := delivery.NewBookHandler(mockUsecase, mockService, templates)

	// Test cases
	t.Run("Valid book ID", func(t *testing.T) {
		// Arrange
		mockUsecase.On("FetchBook", 123).Return(&domain.Book{
			Content:  "This is the content of the book.",
			Metadata: domain.Metadata{Title: "Test Title", Author: "Test Author"},
		}, nil)

		req, _ := http.NewRequest("GET", "/books/123", nil)
		rec := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/books/{id}", handler.Show)

		// Act
		router.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Test Title")
		assert.Contains(t, rec.Body.String(), "Test Author")
		assert.Contains(t, rec.Body.String(), "This is the content of the book.")
	})
}
