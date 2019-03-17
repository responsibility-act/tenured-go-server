package protocol

type AuthHeader struct {
	Module     string            `json:"module"`
	Address    string            `json:"address"`
	Attributes map[string]string `json:"attributes"`
}

func (this *AuthHeader) AddAttributes(key, value string) {
	this.Attributes[key] = value
}
