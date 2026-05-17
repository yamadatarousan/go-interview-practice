// Package main contains the implementation for Challenge 9: RESTful Book Management API
package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// --- Domain ---

// Book represents a book in the database
type Book struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	PublishedYear int    `json:"published_year"`
	ISBN          string `json:"isbn"`
	Description   string `json:"description"`
}

// --- Repository ---

// BookRepository defines the operations for book data access
type BookRepository interface {
	GetAll() ([]*Book, error)
	GetByID(id string) (*Book, error)
	Create(book *Book) error
	Update(id string, book *Book) error
	Delete(id string) error
	SearchByAuthor(author string) ([]*Book, error)
	SearchByTitle(title string) ([]*Book, error)
}

// inMemoryBookRepo implements BookRepository using in-memory storage
type inMemoryBookRepo struct {
	books  map[string]*Book
	mu     sync.RWMutex
	nextID int
}

// NewInMemoryBookRepository creates a new in-memory book repository
func NewInMemoryBookRepository() BookRepository {
	return &inMemoryBookRepo{
		books:  make(map[string]*Book),
		nextID: 1,
	}
}

// Implement BookRepository methods for inMemoryBookRepo
// ...

func (r *inMemoryBookRepo) GetAll() ([]*Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*Book, 0, len(r.books))
	for _, b := range r.books {
		list = append(list, cloneBook(b))
	}
	return list, nil
}

func (r *inMemoryBookRepo) GetByID(id string) (*Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.books[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return cloneBook(b), nil
}

func (r *inMemoryBookRepo) Create(book *Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if book.ID == "" {
		book.ID = strconv.Itoa(r.nextID)
		r.nextID++
	}
	if _, exists := r.books[book.ID]; exists {
		return errors.New("book with this ID already exists")
	}
	r.books[book.ID] = cloneBook(book)
	return nil
}

func (r *inMemoryBookRepo) Update(id string, book *Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.books[id]; !ok {
		return errors.New("not found")
	}
	book.ID = id
	r.books[id] = cloneBook(book)
	return nil
}

func (r *inMemoryBookRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.books[id]; !ok {
		return errors.New("not found")
	}
	delete(r.books, id)
	return nil
}

func (r *inMemoryBookRepo) SearchByAuthor(author string) ([]*Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	author = strings.TrimSpace(strings.ToLower(author))
	var res []*Book
	for _, b := range r.books {
		if strings.Contains(strings.ToLower(b.Author), author) {
			res = append(res, cloneBook(b))
		}
	}
	return res, nil
}

func (r *inMemoryBookRepo) SearchByTitle(title string) ([]*Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	title = strings.TrimSpace(strings.ToLower(title))
	var res []*Book
	for _, b := range r.books {
		if strings.Contains(strings.ToLower(b.Title), title) {
			res = append(res, cloneBook(b))
		}
	}
	return res, nil
}

func cloneBook(b *Book) *Book {
	c := *b
	return &c
}

// --- Service ---

// BookService defines the business logic for book operations
type BookService interface {
	GetAllBooks() ([]*Book, error)
	GetBookByID(id string) (*Book, error)
	CreateBook(book *Book) error
	UpdateBook(id string, book *Book) error
	DeleteBook(id string) error
	SearchBooksByAuthor(author string) ([]*Book, error)
	SearchBooksByTitle(title string) ([]*Book, error)
}

// bookService implements BookService
type bookService struct {
	repo BookRepository
}

// NewBookService creates a new book service
func NewBookService(repo BookRepository) BookService {
	return &bookService{
		repo: repo,
	}
}

func validateBook(b *Book) error {
	if strings.TrimSpace(b.Title) == "" || strings.TrimSpace(b.Author) == "" {
		return errors.New("title and author are required")
	}
	if b.PublishedYear < 0 {
		return errors.New("published year must be non-negative")
	}
	return nil
}

// Implement BookService methods for bookService
// ...

func (s *bookService) GetAllBooks() ([]*Book, error)                 { return s.repo.GetAll() }
func (s *bookService) GetBookByID(id string) (*Book, error)          { return s.repo.GetByID(id) }
func (s *bookService) CreateBook(b *Book) error                      { return s.repo.Create(b) }
func (s *bookService) UpdateBook(id string, b *Book) error           { return s.repo.Update(id, b) }
func (s *bookService) DeleteBook(id string) error                    { return s.repo.Delete(id) }
func (s *bookService) SearchBooksByAuthor(a string) ([]*Book, error) { return s.repo.SearchByAuthor(a) }
func (s *bookService) SearchBooksByTitle(t string) ([]*Book, error)  { return s.repo.SearchByTitle(t) }

// --- Handler ---

// BookHandler handles HTTP requests for book operations
type BookHandler struct {
	Service BookService
}

// NewBookHandler creates a new book handler
func NewBookHandler(service BookService) *BookHandler {
	return &BookHandler{Service: service}
}

func (h *BookHandler) handleGetByID(w http.ResponseWriter, r *http.Request, id string) {
	book, err := h.Service.GetBookByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, errors.New("book not found"))
		return
	}
	respondJSON(w, http.StatusOK, book)
}

func (h *BookHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		respondError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	if err := validateBook(&b); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.Service.CreateBook(&b); err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, http.StatusCreated, b)
}

func (h *BookHandler) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		respondError(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	if err := validateBook(&b); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.Service.UpdateBook(id, &b); err != nil {
		respondError(w, http.StatusNotFound, errors.New("book not found"))
		return
	}
	b.ID = id
	respondJSON(w, http.StatusOK, b)
}

func (h *BookHandler) handleDelete(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.Service.DeleteBook(id); err != nil {
		respondError(w, http.StatusNotFound, errors.New("book not found"))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "book deleted"})
}

func (h *BookHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")
	title := r.URL.Query().Get("title")
	var (
		books []*Book
		err   error
	)
	if author != "" {
		books, err = h.Service.SearchBooksByAuthor(author)
	} else if title != "" {
		books, err = h.Service.SearchBooksByTitle(title)
	} else {
		respondError(w, http.StatusBadRequest, errors.New("missing search query"))
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, books)
}

func respondJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, code int, err error) {
	respondJSON(w, code, map[string]string{"error": err.Error()})
}

func (h *BookHandler) HandleBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1) /api/books/search?author=... or?title=...
	if strings.HasPrefix(r.URL.Path, "/api/books/search") {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
			return
		}
		author := r.URL.Query().Get("author")
		title := r.URL.Query().Get("title")
		var books []*Book
		var err error
		if author != "" {
			books, err = h.Service.SearchBooksByAuthor(author)
		} else if title != "" {
			books, err = h.Service.SearchBooksByTitle(title)
		} else {
			respondError(w, http.StatusBadRequest, errors.New("author or title required"))
			return
		}
		if err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}
		respondJSON(w, http.StatusOK, books)
		return
	}

	// 2) /api/books/{id}
	if strings.HasPrefix(r.URL.Path, "/api/books/") {
		id := strings.TrimPrefix(r.URL.Path, "/api/books/")
		id = strings.Split(id, "/")[0] // 余計な / を除去
		if id == "" {
			respondError(w, http.StatusNotFound, errors.New("book not found"))
			return
		}

		switch r.Method {
		case http.MethodGet:
			b, err := h.Service.GetBookByID(id)
			if err != nil {
				respondError(w, http.StatusNotFound, errors.New("book not found"))
				return
			}
			respondJSON(w, http.StatusOK, b)
			return
		case http.MethodPut:
			var b Book
			if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
				respondError(w, http.StatusBadRequest, errors.New("invalid json"))
				return
			}
			if err := validateBook(&b); err != nil {
				respondError(w, http.StatusBadRequest, err)
				return
			}
			if err := h.Service.UpdateBook(id, &b); err != nil {
				respondError(w, http.StatusNotFound, errors.New("book not found"))
				return
			}
			b.ID = id
			respondJSON(w, http.StatusOK, b)
			return
		case http.MethodDelete:
			if err := h.Service.DeleteBook(id); err != nil {
				respondError(w, http.StatusNotFound, errors.New("book not found"))
				return
			}
			respondJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
			return
		default:
			respondError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
			return
		}
	}

	// 3) /api/books （一覧・作成）
	switch r.Method {
	case http.MethodGet:
		books, _ := h.Service.GetAllBooks()
		if books == nil {
			books = []*Book{}
		}
		respondJSON(w, http.StatusOK, books)
	case http.MethodPost:
		var b Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			respondError(w, http.StatusBadRequest, errors.New("invalid json"))
			return
		}
		if err := validateBook(&b); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}
		if err := h.Service.CreateBook(&b); err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}
		respondJSON(w, http.StatusCreated, b)
	default:
		respondError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (h *BookHandler) HandleBookByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := ""
	if len(parts) >= 3 {
		id = parts[2]
	}

	switch r.Method {
	case http.MethodGet:
		b, err := h.Service.GetBookByID(id)
		if err != nil {
			respondError(w, http.StatusNotFound, errors.New("book not found"))
			return
		}
		respondJSON(w, http.StatusOK, b)
	case http.MethodPut:
		var b Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			respondError(w, http.StatusBadRequest, errors.New("invalid json"))
			return
		}
		if strings.TrimSpace(b.Title) == "" || strings.TrimSpace(b.Author) == "" {
			respondError(w, http.StatusBadRequest, errors.New("title and author required"))
			return
		}
		if err := h.Service.UpdateBook(id, &b); err != nil {
			respondError(w, http.StatusNotFound, errors.New("book not found"))
			return
		}
		b.ID = id
		respondJSON(w, http.StatusOK, b)
	case http.MethodDelete:
		if err := h.Service.DeleteBook(id); err != nil {
			respondError(w, http.StatusNotFound, errors.New("book not found"))
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
	default:
		respondError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
}

func (h *BookHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	author := r.URL.Query().Get("author")
	title := r.URL.Query().Get("title")

	var books []*Book
	var err error
	if author != "" {
		books, err = h.Service.SearchBooksByAuthor(author)
	} else if title != "" {
		books, err = h.Service.SearchBooksByTitle(title)
	} else {
		respondError(w, http.StatusBadRequest, errors.New("author or title required"))
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, books)
}
