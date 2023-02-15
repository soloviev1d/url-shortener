package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

type Server struct {
	addr string
}

var (
	database *pgx.Conn
	permId   = 1
	err      error
)

func NewServer(addr string) (*Server, error) {
	database, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	firstNotNull := database.QueryRow(context.Background(), "select id from urls.urls where original_url is not null")
	firstNotNull.Scan(&permId)
	//fill db
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/shorten", shortenUrlHandler)
	http.HandleFunc("/", getUrlHandler)

	return &Server{
		addr: addr,
	}, nil
}

func (s *Server) ListenAndServe() error {
	if err := http.ListenAndServe(s.addr, nil); err != nil {
		return err
	}
	return nil
}

func shortenUrlHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")

	var returnUrl string
	err := database.QueryRow(context.Background(), "select shortened from urls.urls where original_url=$1", url).Scan(&returnUrl)
	if err != nil && err != pgx.ErrNoRows {
		http.Error(w, "failed to access the database", http.StatusInternalServerError)
	}
	if len(returnUrl) > 0 {
		fmt.Fprintf(w, "url: https://url-shortener.soloviev1d.repl.co/%s", returnUrl)
		return
	}

	_, err = database.Exec(context.Background(), "update urls.urls set original_url=$1 where id=$2;", url, permId)
	if err != nil {
		http.Error(w, "failed to update database", http.StatusInternalServerError)
	}

	err = database.QueryRow(context.Background(), "select shortened from urls.urls where id=$1", permId).Scan(&returnUrl)
	fmt.Println(returnUrl, permId)
	if err != nil {
		http.Error(w, "failed to retrieve shortened url", http.StatusInternalServerError)
	}

	if permId <= 720 {
		permId++
	} else {
		permId = 1
	}

	fmt.Fprintf(w, "url: https://url-shortener.soloviev1d.repl.co/%s", returnUrl)

}

func getUrlHandler(w http.ResponseWriter, r *http.Request) {
	var url string
	err := database.QueryRow(context.Background(), "select original_url from urls.urls where shortened=$1", r.URL.Path[1:]).Scan(&url)
	switch err {
	case nil:
		http.Redirect(w, r, url, http.StatusSeeOther)
	case pgx.ErrNoRows:
		http.NotFound(w, r)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
