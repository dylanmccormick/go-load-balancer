package balancer

type Backend struct {
	ConnType string
	Host     string
	Port     string
	Healthy  bool
}
