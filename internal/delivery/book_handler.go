package delivery

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
	Usecase   usecase.IBookUsecase
	Service   service.IAnalysisService
	Templates *template.Template
	Logger    *service.Logger
}

func NewBookHandler(usecase usecase.IBookUsecase, analysis service.IAnalysisService, tmpl *template.Template) *BookHandler {
	logger := service.NewLogger("[BookHandler]")
	return &BookHandler{Usecase: usecase, Service: analysis, Templates: tmpl, Logger: logger}
}

func (h *BookHandler) Index(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "index.html", "", "", "")
}

func (h *BookHandler) Show(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gutenbergID := vars["id"]

	h.Logger.SetTags(fmt.Sprintf("[book-%s]", gutenbergID))

	id, err := strconv.Atoi(gutenbergID)
	if err != nil {
		h.Logger.LogError("Failed to parse gutenbergID", err)
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
	gutenbergID := vars["id"]

	h.Logger.SetTags(fmt.Sprintf("[book-%s]", gutenbergID))

	id, err := strconv.Atoi(gutenbergID)
	if err != nil {
		h.Logger.LogError("Failed to parse gutenbergID", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	book, err := h.Usecase.FetchBook(id)
	if err != nil {
		h.Logger.LogError("Failed to fetch book", err)
		http.Error(w, "Failed to fetch book", http.StatusNotFound)
		return
	}

	err = h.Service.StreamTextAnalysis(w, r, book.Content)
	if err != nil {
		http.Error(w, "Failed to stream analysis: "+err.Error(), http.StatusInternalServerError)
	}
}
