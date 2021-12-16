package nunit3

import "encoding/xml"

type Reason struct {
	XMLName xml.Name `xml:"reason"`
	Message *reasonMessage
}

func NewReason(msg string) *Reason {
	f := new(Reason)
	f.Message = &reasonMessage{Message: msg}
	return f
}

type reasonMessage struct {
	XMLName xml.Name `xml:"message"`
	Message string   `xml:",cdata"`
}
