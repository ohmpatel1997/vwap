package producers

//go:generate mockery --name Producer --output ../../pkg/mocks/producers --outpkg producers
type Producer interface {
	VWAP() VWAP
}

type producer struct {
	vwap VWAP
}

func NewProducer() Producer {
	return &producer{
		vwap: newVWAP(),
	}
}

func (m *producer) VWAP() VWAP {
	return m.vwap
}
