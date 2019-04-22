package protocol

import "fmt"

type AuthHeader struct {
	Module     string            `json:"module"`
	Address    string            `json:"address"`
	Attributes map[string]string `json:"attributes"`
}

func (this *AuthHeader) AddAttributes(key, value string) {
	this.Attributes[key] = value
}

func (this *AuthHeader) String() string {
	return fmt.Sprintf("AuthHeader{module=%s, address=%s, attrs=%v}",
		this.Module, this.Address, this.Attributes)
}
