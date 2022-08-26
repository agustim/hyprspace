package webserver

import (
	"encoding/json"
	"log"

	"net/http"

	"github.com/hyprspace/hyprspace/config"
	"github.com/julienschmidt/httprouter"
)

type Server struct {
	router *httprouter.Router
	cfg    *config.Config
}

// Create a http server to read and edit the config file
func CreateServer(cfg *config.Config) {
	// Create a new web server
	server := &Server{cfg: cfg}
	server.Init()
	server.Run()
}

func (s *Server) Init() {
	s.router = httprouter.New()

	s.router.GET("/", s.Index)
	s.router.GET("/routes", s.Routes)
	s.router.GET("/add-route/:net/:mask/:gateway", s.AddRoute)
	s.router.GET("/remove-route/:net/:mask/:gateway", s.RemoveRoute)
}

func (s *Server) Run() {
	log.Fatal(http.ListenAndServe(":8080", s.router))
}

func (s *Server) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var PrivateKey string
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	PrivateKey = s.cfg.Interface.PrivateKey
	s.cfg.Interface.PrivateKey = "<hidden>"
	jsonResp, err := json.Marshal(s.cfg)
	s.cfg.Interface.PrivateKey = PrivateKey
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func (s *Server) Routes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(s.cfg.Routes)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func (s *Server) AddRoute(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	network := ps.ByName("net")
	mask := ps.ByName("mask")
	cidr := network + "/" + mask
	gateway := ps.ByName("gateway")
	s.cfg.Routes[cidr] = config.Route{IP: gateway}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(s.cfg.Routes)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func (s *Server) RemoveRoute(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	network := ps.ByName("net")
	mask := ps.ByName("mask")
	cidr := network + "/" + mask
	delete(s.cfg.Routes, cidr)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(s.cfg.Routes)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}
