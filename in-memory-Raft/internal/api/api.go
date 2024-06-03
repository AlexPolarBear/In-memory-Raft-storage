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

func New(addr string, store *services.InMemoryStore) *StorageController {
	return &StorageController{
		addr:  addr,
		store: store,
	}
}

func (sc *StorageController) Start() error {
	r := mux.NewRouter()
	r.HandleFunc("/keys/{key}", sc.HandleGetKey).Methods("GET")
	r.HandleFunc("/keys", sc.HandlePostKey).Methods("POST")
	r.HandleFunc("/keys/{key}", sc.HandleDeleteKey).Methods("DELETE")
	r.HandleFunc("/join", sc.HandleJoin).Methods("POST")
	// r.PathPrefix("/").Handler(httpSwagger.WrapHandler)

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

func (sc *StorageController) HandleGetKey(w http.ResponseWriter, r *http.Request) {
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

func (sc *StorageController) HandlePostKey(w http.ResponseWriter, r *http.Request) {
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

func (sc *StorageController) HandleDeleteKey(w http.ResponseWriter, r *http.Request) {
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
	sc.store.Delete(k)

}

func (sc *StorageController) Addr() net.Addr {
	return sc.ln.Addr()
}
