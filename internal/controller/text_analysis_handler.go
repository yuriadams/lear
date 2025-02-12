package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/yuriadams/lear/internal/usecase"

	"github.com/gorilla/mux"
)

type TextAnalysisHandler struct {
	Usecase *usecase.BookUsecase
}

func NewTextAnalysisHandler(usecase *usecase.BookUsecase) *TextAnalysisHandler {
	return &TextAnalysisHandler{Usecase: usecase}
}

func (h *TextAnalysisHandler) Post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	analysis, err := h.Usecase.AnalyzeBookContent(id)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Analysis failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(analysis)
}
