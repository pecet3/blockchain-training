package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	PublishDate string `json:"publish_date"`
	ISBN        string `json:"isbn"`
}

type BookCheckout struct {
	BookID       string `json:"book_id"`
	User         string `json:"user"`
	CheckoutDate string `json:"checkout_date"`
	IsGenesis    string `json:"is_genesis"`
}

type Block struct {
	Position  int
	Data      BookCheckout
	TimeStamp string
	Hash      string
	PrevHash  string
}

type Blockchain struct {
	blocks []*Block
}

var blockchain *Blockchain

func newBook(w http.ResponseWriter, r *http.Request) {
	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		w.Write([]byte("could not create a new book"))
		return
	}

	h := md5.New()

	io.WriteString(h, book.ISBN+book.PublishDate)
	book.ID = fmt.Sprintf("%x", h.Sum(nil))

	bytes, err := json.MarshalIndent(book, "", " ")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error MarshalIndent"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkoutItem BookCheckout

	if err := json.NewDecoder(r.Body).Decode(&checkoutItem); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		w.Write([]byte("error json"))
		return
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", getBlockchain)
	mux.HandleFunc("POST /", writeBlock)
	mux.HandleFunc("POST /new", newBook)

	log.Fatal(http.ListenAndServe(":8090", mux))
}
