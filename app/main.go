package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

// | ID (16 bits) | QR (1 bit) | OpCode (4 bits) | AA (1 bit) | TC (1 bit) | RD (1 bit) | RA (1 bit) | Z (3 bits) | RCode (4 bits) |
// |---0---|---1--|-----------------------------2---------------------------------------|------------------3-----------------------|

type DNSMessage struct {
	Data      []byte
	labelsEnd int
}

func New(d []byte) DNSMessage {
	return DNSMessage{
		Data: d,
	}
}

func (m *DNSMessage) GetID() uint16 {
	return binary.BigEndian.Uint16(m.Data[:2])
}

func (m *DNSMessage) SetID(v uint16) {
	binary.BigEndian.PutUint16(m.Data[:2], v)
}

func (m *DNSMessage) GetQR() uint8 {
	return (m.Data[2] & 0x80) >> 7
}

func (m *DNSMessage) SetQR(v bool) {
	if v {
		m.Data[2] |= 0x80
	} else {
		m.Data[2] &= 0x7F
	}
}

func (m *DNSMessage) GetQDCount() uint16 {
	return binary.BigEndian.Uint16(m.Data[4:6])
}

func (m *DNSMessage) SetQDCount(v uint16) {
	binary.BigEndian.PutUint16(m.Data[4:6], v)
}

func (m *DNSMessage) GetANCount() uint16 {
	return binary.BigEndian.Uint16(m.Data[6:8])
}

func (m *DNSMessage) SetANCount(v uint16) {
	binary.BigEndian.PutUint16(m.Data[6:8], v)
}

func (m *DNSMessage) GetQName() string {
	var sb strings.Builder

	p := 12
	for m.Data[p] != 0x00 {
		l := int(m.Data[p]) // How many bytes in this label?
		p += 1              // Move on from length byte

		// Get each char in the label
		for i := 0; i < l; i++ {
			sb.WriteByte(m.Data[p])
			p += 1
		}

		if m.Data[p] != 0x00 {
			sb.WriteByte('.')
		}
	}

	return sb.String()
}

func (m *DNSMessage) SetQName(v string) {
	sts := strings.Split(v, ".")

	p := 12
	for _, st := range sts {
		// Set length of label
		m.Data[p] = byte(len(st))
		p += 1

		// Write each of the chars for the label
		for _, c := range st {
			m.Data[p] = byte(c)
			p += 1
		}
	}

	m.Data[p] = 0x00    // Null terminator
	m.labelsEnd = p + 1 // Move past the null terminator
}

func (m *DNSMessage) GetQType() uint16 {
	return binary.BigEndian.Uint16(m.Data[m.labelsEnd : m.labelsEnd+2])

}

func (m *DNSMessage) SetQType(v uint16) {
	binary.BigEndian.PutUint16(m.Data[m.labelsEnd:m.labelsEnd+2], v)
}

func (m *DNSMessage) GetQClass() uint16 {
	return binary.BigEndian.Uint16(m.Data[m.labelsEnd+2 : m.labelsEnd+4])

}

func (m *DNSMessage) SetQClass(v uint16) {
	binary.BigEndian.PutUint16(m.Data[m.labelsEnd+2:m.labelsEnd+4], v)
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

		request := New(buf[:size])
		response := New(make([]byte, 512))

		response.SetID(request.GetID()) // Response packets must reply with the same ID.
		response.SetQR(true)            // Indicates it's a response
		response.SetQDCount(1)          // Number of questions in the query
		response.SetANCount(1)          // Number of answers in the response

		response.SetQName(request.GetQName()) // Domain name
		response.SetQType(1)                  // A
		response.SetQClass(1)

		fmt.Printf("RedID: %v\n", request.GetID())
		fmt.Printf("ResID: %v\n", response.GetID())
		fmt.Printf("QR: %v\n", response.GetQR())
		fmt.Printf("QDCount: %v\n", response.GetQDCount())
		fmt.Printf("ANCount: %v\n", response.GetANCount())
		fmt.Printf("QName: %v\n", response.GetQName())
		fmt.Printf("QType: %v\n", response.GetQType())
		fmt.Printf("QClass: %v\n", response.GetQClass())
		fmt.Printf("Request Data: %v\n", request.Data[:size])
		fmt.Printf("Response Data: %v\n", response.Data[:50])
		fmt.Printf("=========\n")

		_, err = udpConn.WriteToUDP(response.Data, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
