package main

import (
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const urlChars = ("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// todo: cache will overfill and is subject to easy DoS attack
var cache map[string]*[]byte
var db *sql.DB

func main() {
	cache = make(map[string]*[]byte)
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("create table if not exists store (id text, val blob);")
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe("0.0.0.0:3000", http.HandlerFunc(handler))
	if err != nil {
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			errType, ok := err.(error)
			if ok {
				fmt.Printf("request panic: %w\n", errType)
			} else {
				panic(err)
			}
		}
	}()

	if r.Method == "GET" {
		handleGet(w, r)
	} else if r.Method == "POST" && r.URL.Path == "/" {
		handlePost(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	println("try get:", key)

	// try cache lookup first
	res, ok := cache[key]
	if ok {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(*res)
		if err != nil {
			panic(err)
		}
		return
	}

	rows, err := db.Query("select val from store where id = $1;", key)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	next := rows.Next()
	if !next {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var val []byte
	err = rows.Scan(&val)

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(val)
	if err != nil {
		panic(err)
	}
	cache[key] = &val
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	id := randomString(8)

	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	res, err := db.Exec("insert into store (id, val) values ($1, $2);", id, data)
	if err != nil {
		panic(err)

	}
	if rowsAffected, err := res.RowsAffected(); rowsAffected < 1 {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

func randomString(length int) string {
	if length < 0 {
		panic(fmt.Errorf("randomString: invalid length %d", length))
	}

	val := make([]rune, length)
	for i := 0; i < length; i++ {
		val[i] = rune(urlChars[rand.Intn(len(urlChars))])
	}
	return string(val)
}
