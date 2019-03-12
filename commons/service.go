package commons

type Service interface {
	Start() error
	Shutdown()
}
