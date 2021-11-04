package proto

import "net"

type SimpleAddr struct {
	NetworkString string `json:"Network"`
	AddressString string `json:"String"`
}

func NewSimpleAddr(addr net.Addr) *SimpleAddr {
	return &SimpleAddr{
		NetworkString: addr.Network(),
		AddressString: addr.String(),
	}
}

func (s SimpleAddr) Network() string {
	return s.NetworkString

}

func (s SimpleAddr) String() string {
	return s.AddressString
}
