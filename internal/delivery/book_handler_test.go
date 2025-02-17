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

type MockBookUsecase struct {
	mock.Mock
}

func (m *MockBookUsecase) FetchBook(id int) (*domain.Book, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Book), args.Error(1)
}

func (m *MockBookUsecase) FetchAllBooks() ([]domain.Book, error) {
	args := m.Called()
	return args.Get(0).([]domain.Book), args.Error(1)
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

	_, err = tmpl.New("index.html").Parse(`
		<h1>Welcome to Project King Lear Explorer</h1>
		<ul>
			{{range .Books}}
				<li>
					<strong>Title:</strong> {{.Title}}<br>
					<strong>Author:</strong> {{.Author}}<br>
					<strong>ID:</strong> {{.GutenbergID}}
				</li>
			{{end}}
		</ul>
	`)
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

func TestBookHandler_Index(t *testing.T) {
	mockUsecase := new(MockBookUsecase)
	mockService := new(MockAnalysisService)
	templates := createTestTemplates()

	handler := delivery.NewBookHandler(mockUsecase, mockService, templates)

	t.Run("Listing books", func(t *testing.T) {
		mockUsecase.On("FetchAllBooks").Return([]domain.Book{
			{
				GutenbergID: 1,
				Metadata:    domain.Metadata{Title: "Test Title 1", Author: "Author 1"},
			},
			{
				GutenbergID: 2,
				Metadata:    domain.Metadata{Title: "Test Title 2", Author: "Author 2"},
			},
		}, nil)

		req, _ := http.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/", handler.Index)

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		body := rec.Body.String()
		assert.Contains(t, body, "Test Title 1")
		assert.Contains(t, body, "Author 1")
		assert.Contains(t, body, "Test Title 2")
		assert.Contains(t, body, "Author 2")
	})
}

func TestBookHandler_Show(t *testing.T) {
	mockUsecase := new(MockBookUsecase)
	mockService := new(MockAnalysisService)
	templates := createTestTemplates()

	handler := delivery.NewBookHandler(mockUsecase, mockService, templates)

	t.Run("Valid book ID", func(t *testing.T) {

		mockUsecase.On("FetchBook", 123).Return(&domain.Book{
			Content:  "This is the content of the book.",
			Metadata: domain.Metadata{Title: "Test Title", Author: "Test Author"},
		}, nil)

		req, _ := http.NewRequest("GET", "/books/123", nil)
		rec := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/books/{id}", handler.Show)

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Test Title")
		assert.Contains(t, rec.Body.String(), "Test Author")
		assert.Contains(t, rec.Body.String(), "This is the content of the book.")
	})
}

func TestBookHandler_StreamAnalysis(t *testing.T) {
	mockUsecase := new(MockBookUsecase)
	mockService := new(MockAnalysisService)
	templates := createTestTemplates()

	handler := delivery.NewBookHandler(mockUsecase, mockService, templates)

	t.Run("Stream analysis with valid book ID", func(t *testing.T) {
		mockUsecase.On("FetchBook", 123).Return(&domain.Book{
			Content:  "This is the content of the book.",
			Metadata: domain.Metadata{Title: "Test Title", Author: "Test Author"},
		}, nil)

		mockService.On("StreamTextAnalysis", mock.Anything, mock.Anything, "This is the content of the book.").Return(nil)

		req, _ := http.NewRequest("GET", "/books/123/analyze", nil)
		rec := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/books/{id}/analyze", handler.StreamAnalysis)

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		mockUsecase.AssertCalled(t, "FetchBook", 123)
		mockService.AssertCalled(t, "StreamTextAnalysis", mock.Anything, mock.Anything, "This is the content of the book.")
	})

	t.Run("Stream analysis with invalid book ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/books/invalid/analyze", nil)
		rec := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/books/{id}/analyze", handler.StreamAnalysis)

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		mockUsecase.AssertNotCalled(t, "FetchBook")
		mockService.AssertNotCalled(t, "StreamTextAnalysis")
	})
}
