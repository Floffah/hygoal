package network

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"hygoal/internal/protocol"
	"log"
	"net"
	"os"

	"github.com/quic-go/quic-go"
)

func StartQuicServer() error {
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: 5520,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		return err
	}
	fmt.Println("QUIC server listening on port 5520")

	transport := quic.Transport{
		Conn: udpConn,
	}

	cert, err := NewSelfSignedServerCert("localhost")
	if err != nil {
		log.Fatalf("generating self-signed cert: %v", err)
	}

	tlsConf := NewQUICServerTLSConfig(cert)

	listener, err := transport.Listen(tlsConf, nil)
	if err != nil {
		return err
	}
	fmt.Println("QUIC server started")

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			if conn != nil {
				conn.CloseWithError(0, "error accepting connection")
			}
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn *quic.Conn) {
	fmt.Println("New connection accepted")
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			conn.CloseWithError(0, "error accepting stream")
			return
		}

		for {
			// packet length
			var packetLen uint32
			lenBuf := make([]byte, 4)
			_, err = stream.Read(lenBuf)
			if err != nil {
				break
			}
			lenReader := bytes.NewReader(lenBuf)
			err = binary.Read(lenReader, binary.LittleEndian, &packetLen)
			if err != nil {
				log.Printf("error reading packet length: %v", err)
				break
			}

			// packet data
			packetData := make([]byte, packetLen+4)
			_, err = stream.Read(packetData)
			if err != nil {
				log.Printf("error reading packet data: %v", err)
				break
			}

			// decode packet ID
			var packetID uint32
			idBuf := bytes.NewReader(packetData[:4])
			err = binary.Read(idBuf, binary.LittleEndian, &packetID)
			if err != nil {
				log.Printf("error reading packet ID: %v", err)
				continue
			}

			packet, err := protocol.DecodeByID(packetID, packetData[4:])
			if err != nil {
				log.Printf("error decoding packet ID %d: %v", packetID, err)
				continue
			}

			log.Printf("Received packet ith ID %d: %+v", packetID, packet)
		}

		//err = debug_writeStream(stream)
		//if err != nil {
		//	log.Printf("error handling stream: %v", err)
		//}

		stream.Close()
	}
}

func debug_writeStream(stream quic.Stream) error {
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil {
		stream.Close()
		return err
	}

	// write out to file
	f, err := os.OpenFile("received_data.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stream.Close()
		return err
	}

	if _, err := f.Write(buf[:n]); err != nil {
		f.Close()
		stream.Close()
		return err
	}
	f.Close()
	stream.Close()
	return nil
}
