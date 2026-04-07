package jwt

type Jwt interface {
	Generate(payload Payload) (*Token, error)
	Parse(finalizedToken string) (*Payload, error)
}
