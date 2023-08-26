package proto98

import (
	"context"
	"fmt"
	"net"

	"github.com/hyprspace/hyprspace/debugpacket"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

func Handler(ctx context.Context, myHost host.Host, destPeer peer.ID, pkt []byte) (err error) {
	// Ha arribat un packet de proto98 i hem de tractar-lo.
	// Primer determinar l'origen i el destí

	// Validar quina commanda és en el byte 21
	// Si és 0x01, és un ping
	switch pkt[21] {
	case 0x01:
		// Ping
		// Enviar un pong
		// Si és 0x02, és un pong
		fmt.Printf("Ping received %s -> %s\n", net.IP(pkt[12:16]).String(), net.IP(pkt[16:20]).String())
		debugpacket.Dump(pkt)
		Pong(ctx, myHost, destPeer, pkt[12:16], pkt[16:20])
	case 0x02:
		// Pong
		fmt.Println("Pong received")
	}
	return nil
}

func GetInformation(pkt []byte) (srcIP string, dstIP string, err error) {
	// Get the source IP
	srcIP = net.IP(pkt[12:16]).String()
	// Get the destination IP
	dstIP = net.IP(pkt[16:20]).String()
	return
}
