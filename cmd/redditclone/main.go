package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"redditclone/internal/controller"
	"redditclone/internal/pkg/auth"
	"redditclone/internal/repository"
)

type Post struct {
	Id string
}

func main() {
	staticDir := "./static"

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_secret_key_change_in_prod"
	}
	tokenManager := auth.NewTokenManager(jwtSecret)

	userRepo := repository.NewMemoryUserRepository()
	postRepo := repository.NewMemoryPostRepository()

	userHandler := controller.NewUserHandler(userRepo, tokenManager)
	postHandler := controller.NewPostHandler(postRepo)

	mux := newMux(staticDir, postHandler, userHandler, tokenManager)

	// Logging the method and path of each request
	handler := loggingMiddleware(mux)

	addr := ":8080"
	log.Printf("Server running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server startup error: %v", err)
	}
}

func newMux(staticDir string, postHandler *controller.PostHandler, userHandler *controller.UserHandler, tokenManager *auth.TokenManager) *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "html", "favicon.ico"))
	})
	mux.HandleFunc("GET /manifest.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "html", "manifest.json"))
	})

	// Authentication routes
	mux.HandleFunc("POST /api/register", userHandler.Register)
	mux.HandleFunc("POST /api/login", userHandler.Login)

	// Post routes
	mux.HandleFunc("GET /api/posts", postHandler.ListPosts)
	mux.HandleFunc("GET /api/posts/{$}", postHandler.ListPosts)
	mux.HandleFunc("GET /api/post/{postId}", postHandler.GetPost)
	mux.HandleFunc("GET /api/posts/{category}", postHandler.ListPostsByCategory)

	// Protected routes for post creation
	protected := controller.AuthMiddleware(tokenManager)
	mux.HandleFunc("POST /api/posts", protected(postHandler.CreatePost))
	mux.HandleFunc("POST /api/posts/{$}", protected(postHandler.CreatePost))

	// Protected routes for comment creation
	mux.HandleFunc("POST /api/post/{postId}", protected(postHandler.CreateComment))
	mux.HandleFunc("DELETE /api/post/{postId}/{commentId}", protected(postHandler.DeleteComment))

	mux.Handle("/api", http.NotFoundHandler())
	mux.Handle("/api/", http.NotFoundHandler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		http.ServeFile(w, r, filepath.Join(staticDir, "html", "index.html"))
	})

	return mux
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
