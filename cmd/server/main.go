package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	auth "signet"

	_ "github.com/mattn/go-sqlite3"
)

type app struct {
	db         *sql.DB
	chainID    *big.Int
	privateKey *rsa.PrivateKey
	pubPEM     string
	nonces     map[string]string
	mu         sync.Mutex
}

type signinRequest struct {
	Address            string `json:"address"`
	Nonce              string `json:"nonce"`
	Signature          string `json:"signature"`
	WalletAddressField string `json:"walletAddressField"`
}

type registerRequest struct {
	Address            string `json:"address"`
	Nonce              string `json:"nonce"`
	Signature          string `json:"signature"`
	WalletAddressField string `json:"walletAddressField"`
	EncryptedUsername  string `json:"encryptedUsername"`
}

func main() {
	dbPath := "signet.db"
	if env := os.Getenv("SQLITE_PATH"); env != "" {
		dbPath = env
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initSchema(db); err != nil {
		log.Fatal(err)
	}

	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(&pk.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))

	a := &app{db: db, chainID: big.NewInt(1), privateKey: pk, pubPEM: pubPEM, nonces: map[string]string{}}

	http.HandleFunc("/api/public-key", a.handlePublicKey)
	http.HandleFunc("/api/challenge", a.handleChallenge)
	http.HandleFunc("/api/signin", a.handleSignIn)
	http.HandleFunc("/api/register", a.handleRegister)
	http.Handle("/", http.FileServer(http.Dir(".")))

	log.Println("server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func initSchema(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS users (
		wallet_address TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

func (a *app) handlePublicKey(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"publicKeyPem": a.pubPEM})
}

func (a *app) handleChallenge(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "address required"})
		return
	}

	nonce, err := randomNonce(16)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	a.mu.Lock()
	a.nonces[address] = nonce
	a.mu.Unlock()

	exists, _, err := findUserByAddress(a.db, address)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"nonce": nonce, "userExists": exists})
}

func (a *app) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var req signinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if !a.checkNonce(req.Address, req.Nonce) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid nonce"})
		return
	}

	ok, err := auth.VerifySignature(req.Address, auth.LOGIN_MESSAGE(), req.Nonce, req.WalletAddressField, req.Signature, a.chainID)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
		return
	}

	exists, username, err := findUserByAddress(a.db, req.Address)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"code": "user_not_exists", "error": "user not exists"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "username": username})
}

func (a *app) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if !a.checkNonce(req.Address, req.Nonce) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid nonce"})
		return
	}

	ok, err := auth.VerifySignature(req.Address, auth.LOGIN_MESSAGE(), req.Nonce, req.WalletAddressField, req.Signature, a.chainID)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
		return
	}

	exists, _, err := findUserByAddress(a.db, req.Address)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists"})
		return
	}

	username, err := a.decryptUsername(req.EncryptedUsername)
	if err != nil || username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid encrypted username"})
		return
	}

	if err := createUser(a.db, req.Address, username); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "registered", "username": username})
}

func (a *app) decryptUsername(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, a.privateKey, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (a *app) checkNonce(address, nonce string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	current, ok := a.nonces[address]
	if !ok || current != nonce {
		return false
	}
	delete(a.nonces, address)
	return true
}

func findUserByAddress(db *sql.DB, address string) (bool, string, error) {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE wallet_address = ?", address).Scan(&username)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, username, nil
}

func createUser(db *sql.DB, address, username string) error {
	_, err := db.Exec("INSERT INTO users(wallet_address, username, created_at) VALUES (?, ?, ?)", address, username, time.Now().UTC())
	return err
}

func randomNonce(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", buf), nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
