package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// | ID (16 bits) | QR (1 bit) | OpCode (4 bits) | AA (1 bit) | TC (1 bit) | RD (1 bit) | RA (1 bit) | Z (3 bits) | RCode (4 bits) |
// |---0---|---1--|-----------------------------2---------------------------------------|------------------3-----------------------|

type DNSMessage struct {
	Data      []byte
	QLabelEnd int
	ALabelEnd int
}

func New(d []byte) DNSMessage {
	return DNSMessage{
		Data: d,
	}
}

// --
// Header
// --
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

// --
// Question
// --
func (m *DNSMessage) GetQName() string {
	v, _ := readName(m.Data, 12)
	return v
}

func (m *DNSMessage) SetQName(v string) {
	i := setName(m.Data, 12, v)
	m.QLabelEnd = i
}

func (m *DNSMessage) GetQType() uint16 {
	return binary.BigEndian.Uint16(m.Data[m.QLabelEnd : m.QLabelEnd+2])

}

func (m *DNSMessage) SetQType(v uint16) {
	binary.BigEndian.PutUint16(m.Data[m.QLabelEnd:m.QLabelEnd+2], v)
}

func (m *DNSMessage) GetQClass() uint16 {
	return binary.BigEndian.Uint16(m.Data[m.QLabelEnd+2 : m.QLabelEnd+4])

}

func (m *DNSMessage) SetQClass(v uint16) {
	binary.BigEndian.PutUint16(m.Data[m.QLabelEnd+2:m.QLabelEnd+4], v)
}

// --
// Answer
// --
func (m *DNSMessage) GetAName() string {
	v, _ := readName(m.Data, m.QLabelEnd+4)
	return v
}

func (m *DNSMessage) SetAName(v string) {
	i := setName(m.Data, m.QLabelEnd+4, v)
	m.ALabelEnd = i
}

func (m *DNSMessage) GetAType() uint16 {
	return binary.BigEndian.Uint16(m.Data[m.ALabelEnd : m.ALabelEnd+2])

}

func (m *DNSMessage) SetAType(v uint16) {
	binary.BigEndian.PutUint16(m.Data[m.ALabelEnd:m.ALabelEnd+2], v)
}

func (m *DNSMessage) GetAClass() uint16 {
	return binary.BigEndian.Uint16(m.Data[m.ALabelEnd+2 : m.ALabelEnd+4])

}

func (m *DNSMessage) SetAClass(v uint16) {
	binary.BigEndian.PutUint16(m.Data[m.ALabelEnd+2:m.ALabelEnd+4], v)
}

func (m *DNSMessage) SetTTL(v uint32) {
	binary.BigEndian.PutUint32(m.Data[m.ALabelEnd+4:m.ALabelEnd+8], v)
}

func (m *DNSMessage) SetRDataLength(v uint16) {
	binary.BigEndian.PutUint16(m.Data[m.ALabelEnd+8:m.ALabelEnd+10], v)
}

func (m *DNSMessage) SetRData(ip string) {
	// Can also do
	// v := net.ParseIP(ip).To4()
	// binary.BigEndian.PutUint32(m.Data[m.ALabelEnd+10:m.ALabelEnd+14], binary.BigEndian.Uint32(v))
	net.ParseIP(ip).To4()
	parts := strings.Split(ip, ".")

	var v []byte = make([]byte, 4)
	for i, part := range parts {
		num, _ := strconv.Atoi(part)
		v[i] = byte(num)
	}

	binary.BigEndian.PutUint32(m.Data[m.ALabelEnd+10:m.ALabelEnd+14], binary.BigEndian.Uint32(v))
}

func readName(data []byte, sp int) (string, int) {
	var sb strings.Builder

	p := sp
	for data[p] != 0x00 {
		l := int(data[p]) // How many bytes in this label?
		p += 1            // Move on from length byte

		// Get each char in the label
		for i := 0; i < l; i++ {
			sb.WriteByte(data[p])
			p += 1
		}

		if data[p] != 0x00 {
			sb.WriteByte('.')
		}
	}

	return sb.String(), p
}

func setName(data []byte, sp int, name string) int {
	sts := strings.Split(name, ".")

	p := sp
	for _, st := range sts {
		// Set length of label
		data[p] = byte(len(st))
		p += 1

		// Write each of the chars for the label
		for _, c := range st {
			data[p] = byte(c)
			p += 1
		}
	}

	data[p] = 0x00 // Null terminator
	return p + 1
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

		// Assumes only one question
		response.SetQName(request.GetQName()) // Domain name
		response.SetQType(1)                  // A
		response.SetQClass(1)

		// Assumes only one answer
		response.SetAName(request.GetQName())
		response.SetAType(1)  // A
		response.SetAClass(1) // IN
		response.SetTTL(60)
		response.SetRDataLength(4)
		response.SetRData("8.8.8.8") // Google DNS IP address

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
