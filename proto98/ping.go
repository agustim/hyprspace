package proto98

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/hyprspace/hyprspace/p2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Prepare ping in proto98
func Ping(ctx context.Context, host host.Host, ipSrc string, ipDest string, peerTable map[string]peer.ID) (err error) {
	fmt.Println("Ping proto98")
	// Prepare a packet IPv4 with protocol is 0x98
	// and destination is the peer we want to ping
	// and source is the IP of the interface
	// Rest parameters are set to 0

	fmt.Println(net.ParseIP(ipDest).To4())
	destIP := net.ParseIP(ipDest)
	// Is another peer?
	peerDest, ok := peerTable[ipDest]
	if !ok {
		err = errors.New("the ip is not a peer")
		return err
	}
	ownerIP, _, err := net.ParseCIDR(ipSrc)
	if err != nil {
		err = errors.New("error happened in parsecidr")
		return err
	}
	packet := PingPacket(ownerIP.To4(), destIP.To4())
	SendMessagePeer(ctx, host, peerDest, packet)
	return nil
}

func BasicPacket(ipSrc []byte, ipDest []byte) (pkt []byte) {
	var packet = make([]byte, 22)
	packet[0] = 0x45
	packet[9] = 0x98
	copy(packet[12:16], ipSrc)
	copy(packet[16:20], ipDest)
	return packet
}

func PingPacket(ipSrc []byte, ipDest []byte) (pkt []byte) {
	packet := BasicPacket(ipSrc, ipDest)
	packet[21] = 0x01 // PING command
	return packet
}

func Pong(ctx context.Context, host host.Host, destPeer peer.ID, ipSrc []byte, ipDest []byte) (err error) {
	packet := PongPacket(ipSrc, ipDest)
	SendMessagePeer(ctx, host, destPeer, packet)
	return
}

func PongPacket(ipSrc []byte, ipDest []byte) (pkt []byte) {
	packet := BasicPacket(ipSrc, ipDest)
	packet[21] = 0x02 // PING command
	return packet
}

func SendMessageStream(stream network.Stream, packet []byte) (err error) {

	plen := len(packet)
	// Write packet length
	err = binary.Write(stream, binary.LittleEndian, uint16(plen))
	if err != nil {
		stream.Close()
	}
	// Write the packet
	_, err = stream.Write(packet[:plen])
	if err != nil {
		stream.Close()
	}
	return
}

func SendMessagePeer(ctx context.Context, self host.Host, peerDest peer.ID, packet []byte) (err error) {
	fmt.Println("[+] New Proto98 Packet from", peerDest.String())

	stream, err := self.NewStream(ctx, peerDest, p2p.Protocol)
	if err != nil {
		fmt.Println("Error happened in NewStream: " + err.Error())
		return
	}
	fmt.Println("Send packet to", peerDest.String())
	return SendMessageStream(stream, packet)
}
