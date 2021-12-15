package nunit3

import "encoding/xml"

type Failure struct {
	XMLName xml.Name `xml:"failure"`
	Message *failureMessage
}

func NewFailure(msg string) *Failure {
	f := new(Failure)
	f.Message = &failureMessage{Message: msg}
	return f
}

type failureMessage struct {
	XMLName xml.Name `xml:"message"`
	Message string   `xml:",cdata"`
}
