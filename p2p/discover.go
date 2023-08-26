package p2p

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hyprspace/hyprspace/config"
	"github.com/hyprspace/hyprspace/state"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Discover starts up a DHT based discovery system finding and adding nodes with the same rendezvous string.
func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, peerTable map[string]peer.ID, i string) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	verbose := ctx.Value(config.WithVerbose) != nil
	if verbose {
		fmt.Println("[+] Starting Discover thread")
	}

	s := make(state.ConnectionState)

	go func() {
		err := <-dht.RefreshRoutingTable()
		if err != nil {
			fmt.Printf("[!] Error Refreshing Routing Table: %v\n", err)
		} else {
			fmt.Println("[+] DHT Routing Table refreshed")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for ip, id := range peerTable {
				s[ip] = h.Network().Connectedness(id) == network.Connected
				if !s[ip] {
					addrs, err := dht.FindPeer(ctx, id)
					if err != nil {
						if verbose {
							fmt.Printf("[!] Couldn't find Peer(%s): %v\n", id, err)
						}
						continue
					}
					_, err = h.Network().DialPeer(ctx, addrs.ID)
					if err != nil {
						if verbose {
							fmt.Printf("[!] Couldn't dial Peer(%s): %v\n", id, err)
						}
						continue
					}
				}

				if verbose {
					fmt.Printf("[+] Connection to %s is alive\n", ip)
				}
			}

			state.Save(i, s)
		}
	}
}

func PrettyDiscovery(ctx context.Context, node host.Host, peerTable map[string]peer.ID) {
	// Build a temporary map of peers to limit querying to only those
	// not connected.
	tempTable := make(map[string]peer.ID, len(peerTable))
	for ip, id := range peerTable {
		tempTable[ip] = id
	}
	for len(tempTable) > 0 {
		for ip, id := range tempTable {
			stream, err := node.NewStream(ctx, id, Protocol)
			if err != nil && (strings.HasPrefix(err.Error(), "failed to dial") ||
				strings.HasPrefix(err.Error(), "no addresses")) {
				// Attempt to connect to peers slowly when they aren't found.
				time.Sleep(5 * time.Second)
				continue
			}
			if err == nil {
				fmt.Printf("[+] Connection to %s Successful. Network Ready.\n", ip)
				stream.Close()
			}
			delete(tempTable, ip)
		}
	}
}
