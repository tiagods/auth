package entity

type Health struct {
	Status  bool
	Service map[string]bool
}

func NewHealth() Health {
	return Health{
		Service: map[string]bool{},
	}
}
