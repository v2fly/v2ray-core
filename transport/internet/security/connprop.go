package security

type ConnectionApplicationProtocol interface {
	GetConnectionApplicationProtocol() (string, error)
}
