package parser

import "fmt"

const imageURLPrefix = "puzzle"

type ImagePart struct {
	host string
	url  string
}

func (ip ImagePart) String() string {
	return fmt.Sprintf("%s%s", ip.host, ip.url)
}
