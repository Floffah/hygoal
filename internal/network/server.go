package network

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

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

	listener, err := transport.Listen(generateTLSConfig(), nil)
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
		fmt.Println("Accepted stream")

		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil {
			stream.Close()
			continue
		}

		// write out to file
		f, err := os.OpenFile("received_data.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			stream.Close()
			continue
		}

		if _, err := f.Write(buf[:n]); err != nil {
			f.Close()
			stream.Close()
			continue
		}
		f.Close()
		stream.Close()
	}
}

func generateTLSConfig() *tls.Config {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Hygoal"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		panic(err)
	}

	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"hytale/1"},
	}
}
