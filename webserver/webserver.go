package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"net/http"

	"github.com/hyprspace/hyprspace/config"
	"github.com/hyprspace/hyprspace/p2p"
	"github.com/hyprspace/hyprspace/proto98"
	"github.com/hyprspace/hyprspace/tun"
	"github.com/julienschmidt/httprouter"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Server struct {
	router    *httprouter.Router
	ctx       context.Context
	host      host.Host
	dht       *dht.IpfsDHT
	cfg       *config.Config
	tunDev    *tun.TUN
	peerTable map[string]peer.ID
	RevLookup map[string]string
}

// Create a http server to read and edit the config file
func CreateServer(ctx context.Context, host host.Host, dht *dht.IpfsDHT, cfg *config.Config, tunDev *tun.TUN, peersTable map[string]peer.ID, RevLookup map[string]string) {
	// Create a new web server
	server := &Server{ctx: ctx, host: host, dht: dht, cfg: cfg, tunDev: tunDev, peerTable: peersTable, RevLookup: RevLookup}
	server.Init()
	server.Run()
}

func (s *Server) Init() {
	s.router = httprouter.New()

	s.router.GET("/", s.Index)
	s.router.GET("/routes", s.Routes)
	s.router.GET("/add-route/:net/:mask/:gateway", s.AddRoute)
	s.router.GET("/remove-route/:net/:mask/:gateway", s.RemoveRoute)
	s.router.GET("/peers", s.Peers)
	s.router.GET("/add-peer/:ip/:peerid", s.AddPeer)
	s.router.GET("/remove-peer/:ip", s.RemovePeer)
	s.router.GET("/ping98/:ip", s.Ping98)
}

func (s *Server) Run() {
	log.Fatal(http.ListenAndServe("localhost:8080", s.router))
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

func (s *Server) Peers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(s.cfg.Peers)
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
	s.RevLookup[peerid] = ip
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(s.cfg.Peers)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	// Setup P2P Discovery
	go p2p.Discover(s.ctx, s.host, s.dht, s.peerTable, s.cfg.Interface.Name)
	go p2p.PrettyDiscovery(s.ctx, s.host, s.peerTable)
}

func (s *Server) RemovePeer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err error
	ip := ps.ByName("ip")
	if peerid, ok := s.peerTable[ip]; ok {
		delete(s.peerTable, ip)
		delete(s.cfg.Peers, ip)
		delete(s.RevLookup, peerid.String())
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(s.RevLookup)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	// Setup P2P Discovery
	go p2p.Discover(s.ctx, s.host, s.dht, s.peerTable, s.cfg.Interface.Name)
	go p2p.PrettyDiscovery(s.ctx, s.host, s.peerTable)
}

func (s *Server) Ping98(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//var err error
	var retorn = []byte("{\"status\": \"ok\"}")
	ip := ps.ByName("ip")

	// Send the packet to peer
	if err := proto98.Ping(s.ctx, s.host, s.cfg.Interface.Address, ip, s.peerTable); err != nil {
		writeErrorJSON(w, err.Error())
		return
	}
	fmt.Println("Ping sent to", ip)
	w.Write(retorn)
}

// Write error JSON to the response
func writeErrorJSON(w http.ResponseWriter, err string) {
	var retorn = []byte("{\"status\": \"error\", \"message\": \"" + err + "\"}")
	fmt.Println(err)
	w.Write(retorn)
}
