package controller

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yuriadams/lear/internal/service"
	"github.com/yuriadams/lear/internal/usecase"
)

type Page struct {
	Title string
	Body  []byte
}

type BookHandler struct {
	Usecase   *usecase.BookUsecase
	Templates *template.Template
	Logger    *service.Logger
}

func NewBookHandler(usecase *usecase.BookUsecase) *BookHandler {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/index.html",
		"web/templates/show.html",
	))

	logger := service.NewLogger("[BookHandler]")
	return &BookHandler{Usecase: usecase, Templates: tmpl, Logger: logger}
}

func (h *BookHandler) Index(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "index.html", "", "", "")
}

func (h *BookHandler) Show(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookId := vars["id"]

	h.Logger.SetTags(fmt.Sprintf("[book-%s]", bookId))

	id, err := strconv.Atoi(bookId)
	if err != nil {
		h.Logger.LogError("Failed to parse bookID", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	book, err := h.Usecase.FetchBook(id)
	if err != nil {
		h.Logger.LogError("Failed to fetch book", err)
		http.Error(w, "Failed to fetch book", http.StatusNotFound)
		return
	}

	h.renderPage(w, "show.html", book.Content, book.Metadata.Title, book.Metadata.Author)
}

func (h *BookHandler) renderPage(w http.ResponseWriter, page, content, title, author string) {
	var body bytes.Buffer
	h.Templates.ExecuteTemplate(&body, page, map[string]interface{}{
		"Title":   title,
		"Author":  author,
		"Content": content,
	})

	h.Templates.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"Title": "Project King Lear Explorer",
		"Body":  template.HTML(body.String()),
	})
}

func (h *BookHandler) StreamAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookId := vars["id"]

	h.Logger.SetTags(fmt.Sprintf("[book-%s]", bookId))

	id, err := strconv.Atoi(bookId)
	if err != nil {
		h.Logger.LogError("Failed to parse bookID", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	book, err := h.Usecase.FetchBook(id)
	if err != nil {
		h.Logger.LogError("Failed to fetch book", err)
		http.Error(w, "Failed to fetch book", http.StatusNotFound)
		return
	}

	err = service.StreamTextAnalysis(w, r, book.Content)
	if err != nil {
		http.Error(w, "Failed to stream analysis: "+err.Error(), http.StatusInternalServerError)
	}
}
