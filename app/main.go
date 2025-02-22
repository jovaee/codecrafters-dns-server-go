package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

// | ID (16 bits) | QR (1 bit) | OpCode (4 bits) | AA (1 bit) | TC (1 bit) | RD (1 bit) | RA (1 bit) | Z (3 bits) | RCode (4 bits) |
// |---0---|---1--|-----------------------------2---------------------------------------|------------------3-----------------------|

type DNSMessage struct {
	Data []byte
}

func New(d []byte) DNSMessage {
	return DNSMessage{
		Data: d,
	}
}

func (m *DNSMessage) GetID() uint16 {
	return binary.BigEndian.Uint16(m.Data[:2])
}

func (d *DNSMessage) SetID(v uint16) {
	binary.BigEndian.PutUint16(d.Data[:2], v)
}

func (d *DNSMessage) GetQR() uint8 {
	return (d.Data[2] & 0x80) >> 7
}

func (d *DNSMessage) SetQR(v bool) {
	if v {
		d.Data[2] |= 0x80
	} else {
		d.Data[2] &= 0x7F
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		msg := New(buf[:size])

		msg.SetID(1234)
		msg.SetQR(true)

		// fmt.Printf("Byte:  %08b\n", msg.Data[2])
		// fmt.Printf("ID: %v\n", msg.GetID())
		// fmt.Printf("QR: %v\n", msg.GetQR())

		// receivedData := string(buf[:size])
		// fmt.Printf("Received %d bytes from %s: %s\n", size, source, receivedData)

		// Create an empty response
		response := msg.Data

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
