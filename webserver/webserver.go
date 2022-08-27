package webserver

import (
	"context"
	"encoding/json"
	"log"

	"net/http"

	"github.com/hyprspace/hyprspace/config"
	"github.com/hyprspace/hyprspace/p2p"
	"github.com/julienschmidt/httprouter"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

type Server struct {
	router    *httprouter.Router
	ctx       context.Context
	host      host.Host
	dht       *dht.IpfsDHT
	cfg       *config.Config
	peerTable map[string]peer.ID
}

// Create a http server to read and edit the config file
func CreateServer(ctx context.Context, host host.Host, dht *dht.IpfsDHT, cfg *config.Config, peersTable map[string]peer.ID) {
	// Create a new web server
	server := &Server{ctx: ctx, host: host, dht: dht, cfg: cfg, peerTable: peersTable}
	server.Init()
	server.Run()
}

func (s *Server) Init() {
	s.router = httprouter.New()

	s.router.GET("/", s.Index)
	s.router.GET("/routes", s.Routes)
	s.router.GET("/add-route/:net/:mask/:gateway", s.AddRoute)
	s.router.GET("/remove-route/:net/:mask/:gateway", s.RemoveRoute)
	s.router.GET("/add-peer/:ip/:peerid", s.AddPeer)
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

func (s *Server) AddPeer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	ip := ps.ByName("ip")
	peerid := ps.ByName("peerid")
	s.peerTable[ip], err = peer.Decode(peerid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, err := json.Marshal(err)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
	}
	s.cfg.Peers[ip] = config.Peer{ID: peerid}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(s.cfg.Peers)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	// Setup P2P Discovery
	go p2p.Discover(s.ctx, s.host, s.dht, s.peerTable)
	go p2p.PrettyDiscovery(s.ctx, s.host, s.peerTable)
}
