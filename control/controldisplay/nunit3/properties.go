package nunit3

import "encoding/xml"

type Property struct {
	XMLName xml.Name `xml:"property"`
	Key     string   `xml:"key"`
	Value   string   `xml:"value"`
}

func NewProperty(key string, value string) *Property {
	return &Property{
		Key:   key,
		Value: value,
	}
}

type Properties struct {
	XMLName xml.Name    `xml:"properties"`
	Props   []*Property `xml:"properties"`
}

func (properties *Properties) AddProperty(pr *Property) {
	properties.Props = append(properties.Props, pr)
}
