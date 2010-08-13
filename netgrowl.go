package netgrowl

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
	"net"
	"os"
)

type NetGrowlError struct {
	os.ErrorString
}

var (
	ErrRegistered    = &NetGrowlError{"Already registered"}
	ErrNotRegistered = &NetGrowlError{"Not registered"}
)

const (
	DefaultAddress = "localhost:9887"

	ProtocolVersion  = 1
	TypeRegistration = 0
	TypeNotification = 1

	PriorityVeryLow   = -2
	PriorityModerate  = -1
	PriorityNormal    = 0
	PriorityHigh      = 1
	PriorityEmergency = 2
)

type NetGrowl struct {
	addr          string
	application   string
	password      string
	notifications []string

	conn *net.UDPConn
}

func NewNetGrowl(addr string, application string, notifications []string, password string) *NetGrowl {
	return &NetGrowl{
		addr:          addr,
		application:   application,
		password:      password,
		notifications: notifications,
	}
}

func (ng *NetGrowl) Register() (err os.Error) {
	if ng.conn != nil {
		return ErrRegistered
	}

	addr, err := net.ResolveUDPAddr(ng.addr)
	if err != nil {
		return
	}

	ng.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}

	payload := bytes.NewBuffer(nil)

	binary.Write(payload, binary.BigEndian, uint8(ProtocolVersion))
	binary.Write(payload, binary.BigEndian, uint8(TypeRegistration))
	binary.Write(payload, binary.BigEndian, uint16(len(ng.application)))

	if ng.notifications != nil {
		binary.Write(payload, binary.BigEndian, uint8(len(ng.notifications)))
		binary.Write(payload, binary.BigEndian, uint8(len(ng.notifications)))
		payload.WriteString(ng.application)

		defaultIndex := bytes.NewBuffer(nil)
		for i, n := range ng.notifications {
			binary.Write(defaultIndex, binary.BigEndian, uint8(i))
			binary.Write(payload, binary.BigEndian, uint16(len(n)))
			payload.WriteString(n)
		}

		io.Copy(payload, defaultIndex)
	} else {
		binary.Write(payload, binary.BigEndian, uint8(0))
		binary.Write(payload, binary.BigEndian, uint8(0))
		payload.WriteString(ng.application)
	}

	hash := md5.New()
	hash.Write(payload.Bytes())
	if ng.password != "" {
		hash.Write([]byte(ng.password))
	}
	binary.Write(payload, binary.BigEndian, hash.Sum())

	_, err = ng.conn.Write(payload.Bytes())
	return
}

func (ng *NetGrowl) Notify(name string, title string, description string, priority int, sticky bool) (err os.Error) {
	if ng.conn == nil {
		return ErrNotRegistered
	}

	flags := (priority & 0x07) << 1
	if priority < 0 {
		flags |= 0x08
	}
	if sticky {
		flags |= 0x01
	}

	payload := bytes.NewBuffer(nil)

	binary.Write(payload, binary.BigEndian, uint8(ProtocolVersion))
	binary.Write(payload, binary.BigEndian, uint8(TypeNotification))
	payload.WriteByte(uint8(flags))
	payload.WriteByte(uint8(flags >> 8))
	binary.Write(payload, binary.BigEndian, uint16(len(name)))
	binary.Write(payload, binary.BigEndian, uint16(len(title)))
	binary.Write(payload, binary.BigEndian, uint16(len(description)))
	binary.Write(payload, binary.BigEndian, uint16(len(ng.application)))

	payload.WriteString(name)
	payload.WriteString(title)
	payload.WriteString(description)
	payload.WriteString(ng.application)

	hash := md5.New()
	hash.Write(payload.Bytes())
	if ng.password != "" {
		hash.Write([]byte(ng.password))
	}
	binary.Write(payload, binary.BigEndian, hash.Sum())

	_, err = ng.conn.Write(payload.Bytes())
	return
}

func (ng *NetGrowl) Close() (err os.Error) {
	if ng.conn == nil {
		return ErrNotRegistered
	}

	err, ng.conn = ng.conn.Close(), nil
	return
}
