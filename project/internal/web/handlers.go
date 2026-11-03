package web

import (
	"io/fs"
	"net/http"
	"strconv"
	"sync"
)

type Todo struct {
	ID    int
	Title string
}

type Server struct {
	render *Renderer
	mu     sync.Mutex
	todos  []Todo
	nextID int
}

func NewServer() (*Server, error) {
	r, err := NewRenderer()
	if err != nil {
		return nil, err
	}
	return &Server{render: r, nextID: 1}, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	static, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))

	mux.HandleFunc("GET /{$}", s.index)
	mux.HandleFunc("POST /todos", s.createTodo)
	mux.HandleFunc("DELETE /todos/{id}", s.deleteTodo)

	return mux
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	todos := append([]Todo(nil), s.todos...)
	s.mu.Unlock()

	s.render.Page(w, "index.html", map[string]any{"Todos": todos})
}

func (s *Server) createTodo(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	t := Todo{ID: s.nextID, Title: title}
	s.nextID++
	s.todos = append(s.todos, t)
	s.mu.Unlock()

	s.render.Partial(w, "todo_item", t)
}

func (s *Server) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	for i, t := range s.todos {
		if t.ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			break
		}
	}
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK) // empty body + outerHTML swap removes the <li>
}
