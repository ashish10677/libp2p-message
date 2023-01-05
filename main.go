package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"log"
	"time"
)

var nodePrivateKeys = []string{
	"w2Am8Wi5cUJ97R7OVyyxgxMTBt+iIB41ttHI4vDHSxI=",
	"C9T6r5XTmeCKRU4/bskWcI3zeGxsT+5Yg9swysNrji4=",
	"agbFmY1sFZVJC5HkIaFCkMPfXCbsl7etNQKI7HEei54=",
}

var multiAddrs = []string{
	"/ip4/192.168.1.90/tcp/3021/p2p/16Uiu2HAmNUDHh8BsiidtPWosMJTAJfnZM1TRY3MQjsuBi7X94gTa",
	"/ip4/192.168.1.90/tcp/3022/p2p/16Uiu2HAmEnGPLkCtVxf3d9k2NiZ8Nz6D7QeHdvvEFPBVt3XCQj4J",
	"/ip4/192.168.1.90/tcp/3023/p2p/16Uiu2HAmDP1SNR1GcSNZrij2YkkRKBptyvmc4HJe6NgZ5SQP6n5V",
}

const ProtocolID = "/chat/1.0.0"

func createNode(port string) (host.Host, error) {
	privKey, pubKey, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	publicKey, err := pubKey.Raw()
	if err != nil {
		return nil, err
	}
	privateKey, err := privKey.Raw()
	if err != nil {
		return nil, err
	}
	fmt.Println("Node Private Key: ", base64.StdEncoding.EncodeToString(privateKey))
	fmt.Println("Node Public Key: ", base64.StdEncoding.EncodeToString(publicKey))
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)),
		libp2p.Identity(privKey),
	)
	if err != nil {
		return nil, err
	}
	return node, nil
}

var dataChannel = make(chan string)

func handleStream(s network.Stream) {
	log.Println("Got a new stream!")

	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReader(s)

	go readData(rw)
	//go writeData(rw)

}

func readData(rw *bufio.Reader) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			dataChannel <- str
		}

	}
}

func writeData(nodeId peer.ID, rw *bufio.Writer) {
	i := 0
	for {
		i += 1
		sendData := fmt.Sprintf("%s: KG Round%d", nodeId, i)
		fmt.Printf("\x1b[32m%s\x1b[0m ----->\n", sendData)
		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
		if i%20 == 0 {
			time.Sleep(1 * time.Second)
		}
	}
}

func startNode(index int, port string) (host.Host, error) {
	privateKey, err := base64.StdEncoding.DecodeString(nodePrivateKeys[index])
	if err != nil {
		panic(err)
	}
	privKey, err := crypto.UnmarshalSecp256k1PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)),
		libp2p.Identity(privKey),
	)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func connectToPeer(node host.Host, destination string) (network.Stream, error) {
	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//// Add the destination's peer multiaddress in the peerstore.
	//// This will be used during connection and stream creation by libp2p.
	//node.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	var connErr error
	for i := 0; i < 5; i++ {
		fmt.Println("Connecting to: ", info.ID)
		connErr = node.Connect(context.Background(), *info)
		if connErr != nil {
			fmt.Println(connErr)
			fmt.Println("Retrying connection in 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}
	}
	if connErr != nil {
		log.Println(err)
		return nil, err
	}

	var (
		stream    network.Stream
		streamErr error
	)
	for i := 0; i < 5; i++ {
		stream, streamErr = node.NewStream(context.Background(), info.ID, ProtocolID)
		if streamErr != nil {
			fmt.Println(err)
			fmt.Println("Retrying opening stream in 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}
	}
	if streamErr != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Established connection to destination")
	return stream, nil
}

func main() {
	var (
		index int
		port  string
		//destination string
	)

	//CreatePrivatePublicKey()

	flag.IntVar(&index, "i", 0, "node index to run")
	flag.StringVar(&port, "p", "3000", "port to start the node on")
	//flag.StringVar(&destination, "d", "", "Destination multiaddr string")
	flag.Parse()

	node, err := startNode(index, port)
	if err != nil {
		panic(err)
	}
	fmt.Println("NODE STARTED: ", node.ID())
	fmt.Println("MULTIADDR: ", node.Addrs())
	node.SetStreamHandler(ProtocolID, handleStream)
	for i := 0; i < 3; i++ {
		if i == index {
			continue
		}
		stream, err := connectToPeer(node, multiAddrs[i])
		if err != nil {
			return
		}
		writer := bufio.NewWriter(stream)
		go writeData(node.ID(), writer)
	}
	for {
		select {
		case str := <-dataChannel:
			// Red console colour: 	\x1b[31m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[31m%s\x1b[0m<-----\n", str)
		}
	}
}
