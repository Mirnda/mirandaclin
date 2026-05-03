package shared

// Address é embutido em User e Clinic via GORM embedded com prefixo "address_".
type Address struct {
	PostalCode   string `json:"postal_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
}
