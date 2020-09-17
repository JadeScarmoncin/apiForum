package main

import (
	"fmt"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"new-forum/apiForum/api"
	"new-forum/apiForum/middleware"
	"os"
	"path/filepath"
	"strings"
)

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	fmt.Println("Début du forum")

	db, err := gorm.Open(sqlite.Open("test.sqlite"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// defer db.Close()

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/static", filesDir)

	db.AutoMigrate(&api.User{})
	db.AutoMigrate(&api.Discussion{})

	r.Route("/users", func(r chi.Router) {
		r.Get("/", api.GetAllUsers(db))
		r.Get("/{id}", api.GetUser(db))
		r.Post("/", api.CreateUser(db))
	})
	r.Route("/discussions", func(r chi.Router) {
		r.Use(middleware.BasicAuth(db))
		r.Get("/", api.GetAllDiscussions(db))
		r.Get("/{id}", api.GetDiscussion(db))
		r.Delete("/{id}", api.DeleteDiscussion(db))
		r.Post("/", api.CreateDiscussion(db))
		// r.Get("/{id}/messages", api.GetAllMessages)
		// r.Get("/messages/{id}", api.GetMessage)
		// r.Post("/{id}/messages", api.CreateMessage)
		// r.Delete("/messages/{id}", api.DeleteMessage)
	})
	http.ListenAndServe(":"+port, r)

}
