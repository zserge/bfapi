package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgres://postgres:bfpasswd@postgres/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS mem (session text, addr integer, val integer,
		PRIMARY KEY(session, addr), CONSTRAINT sessaddr UNIQUE(session, addr));
	`); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess := r.Header.Get("X-Session")
		n, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		row := db.QueryRow("SELECT val FROM mem WHERE session = $1 AND addr = $2", sess, n)
		if err := row.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		val := 0
		row.Scan(&val)
		fmt.Fprintf(w, "%d", val)
	})
	http.HandleFunc("/inc/", func(w http.ResponseWriter, r *http.Request) {
		sess := r.Header.Get("X-Session")
		n, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/inc/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := db.Exec(`INSERT INTO mem (session, addr, val) VALUES ($1, $2, 1)
				ON CONFLICT (session, addr) DO UPDATE SET val = mem.val + 1;`, sess, n); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.HandleFunc("/dec/", func(w http.ResponseWriter, r *http.Request) {
		sess := r.Header.Get("X-Session")
		n, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/dec/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := db.Exec(`INSERT INTO mem (session, addr, val) VALUES ($1, $2, -1)
				ON CONFLICT (session, addr) DO UPDATE SET val = mem.val - 1;`, sess, n); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
