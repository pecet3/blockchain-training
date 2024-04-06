package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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
	IsGenesis    bool   `json:"is_genesis"`
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

func (b *Block) generateHash() {
	bytes, _ := json.Marshal(b.Data)

	data := string(b.Position) + b.TimeStamp + string(bytes) + b.PrevHash

	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func CreateBlock(prev *Block, data BookCheckout) *Block {
	block := &Block{}
	block.Position = prev.Position + 1
	block.PrevHash = prev.Hash
	block.TimeStamp = time.Now().String()
	block.Data = data
	block.generateHash()

	return block
}

func validBlock(block *Block, prev *Block) bool {
	if prev.Hash != block.PrevHash {
		return false
	}
	if !block.validateHash(block.Hash) {
		return false
	}
	if prev.Position+1 != block.Position {
		return false
	}
	return true
}

func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	if b.Hash != hash {
		return false
	}
	return true
}

func (bc *Blockchain) AddBlock(data BookCheckout) {
	prevBlock := bc.blocks[len(bc.blocks)-1]

	block := CreateBlock(prevBlock, data)

	if validBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	}
}

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

	blockchain.AddBlock(checkoutItem)
}

func GenesisBlock() *Block {
	return CreateBlock(&Block{}, BookCheckout{IsGenesis: true})
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(blockchain.blocks, "", " ")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
	}
	io.WriteString(w, string(bytes))
	w.Write(bytes)
}

func main() {
	blockchain = NewBlockchain()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", getBlockchain)
	mux.HandleFunc("POST /", writeBlock)
	mux.HandleFunc("POST /new", newBook)

	go func() {
		for _, block := range blockchain.blocks {
			fmt.Printf("prev, hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data:%v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
	}()

	log.Fatal(http.ListenAndServe(":8090", mux))
}
