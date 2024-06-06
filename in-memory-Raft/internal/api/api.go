package api

import (
	"encoding/json"
	"inmemoryraft/internal/services"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

type StorageController struct {
	addr string
	ln   net.Listener

	store *services.InMemoryStore
}

func NewInMemoryStore(addr string, store *services.InMemoryStore) *StorageController {
	return &StorageController{
		addr:  addr,
		store: store,
	}
}

func (sc *StorageController) Starter() error {
	r := mux.NewRouter()
	r.HandleFunc("/keys/{key}", sc.HandleGet).Methods("GET")
	r.HandleFunc("/keys", sc.HandlePut).Methods("POST")
	r.HandleFunc("/keys/{key}", sc.HandleDelete).Methods("DELETE")
	r.HandleFunc("/join", sc.HandleJoin).Methods("POST")
	r.HandleFunc("/load-transaction-log", sc.HandleLoadTransactionLog).Methods("GET")
	r.HandleFunc("/save-transaction-log", sc.HandleSaveTransactionLog).Methods("GET")
	r.Handle("/", http.FileServer(http.Dir("configs")))

	server := http.Server{
		Handler: r,
	}

	ln, err := net.Listen("tcp", sc.addr)
	if err != nil {
		return err
	}
	sc.ln = ln

	http.Handle("/", r)

	go func() {
		err := server.Serve(sc.ln)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()

	return nil
}

func (sc *StorageController) Close() {
	sc.ln.Close()
}

func (sc *StorageController) HandleJoin(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := m["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := sc.store.Join(nodeID, remoteAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (sc *StorageController) HandleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	val, err := sc.store.Get(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(map[string]string{key: val})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(b))
}

func (sc *StorageController) HandlePut(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for k, v := range m {
		if err := sc.store.Put(k, v); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (sc *StorageController) HandleDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	k := vars["key"]

	if k == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := sc.store.Delete(k); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (sc *StorageController) Addr() net.Addr {
	return sc.ln.Addr()
}

func (sc *StorageController) HandleLoadTransactionLog(w http.ResponseWriter, r *http.Request) {
	if err := sc.store.LoadTransactionLog(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Transaction log loaded successfully")
}

func (sc *StorageController) HandleSaveTransactionLog(w http.ResponseWriter, r *http.Request) {
	if err := sc.store.SaveTransactionLog(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Transaction log saved successfully")
}
