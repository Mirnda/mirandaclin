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
	Latitude     string `json:"latitude,omitempty"`
	Longitude    string `json:"longitude,omitempty"`
}

// AddressInput é usado em requests HTTP. *string é para distinguir
// campo ausente (nil) de campo presente mas vazio (""), permitindo validate:"required".
type AddressInput struct {
	PostalCode   *string `json:"postal_code" validate:"required"`
	Street       *string `json:"street"`
	Number       *string `json:"number"`
	Complement   *string `json:"complement"`
	Neighborhood *string `json:"neighborhood"`
	City         *string `json:"city" validate:"required"`
	State        *string `json:"state" validate:"required,min=2"`
	Country      *string `json:"country"`
	Latitude     *string `json:"latitude"`
	Longitude    *string `json:"longitude"`
}

func (a AddressInput) ToAddress() Address {
	var address Address

	if a.PostalCode != nil {
		address.PostalCode = *a.PostalCode
	}
	if a.Street != nil {
		address.Street = *a.Street
	}
	if a.Number != nil {
		address.Number = *a.Number
	}
	if a.Complement != nil {
		address.Complement = *a.Complement
	}
	if a.Neighborhood != nil {
		address.Neighborhood = *a.Neighborhood
	}
	if a.City != nil {
		address.City = *a.City
	}
	if a.State != nil {
		address.State = *a.State
	}
	if a.Country != nil {
		address.Country = *a.Country
	}
	if a.Latitude != nil {
		address.Latitude = *a.Latitude
	}
	if a.Longitude != nil {
		address.Longitude = *a.Longitude
	}

	// 	return Address{ //TODO: REMOVER
	// 	PostalCode:   *a.PostalCode,
	// 	Street:       *a.Street,
	// 	Number:       *a.Number,
	// 	Complement:   *a.Complement,
	// 	Neighborhood: *a.Neighborhood,
	// 	City:         *a.City,
	// 	State:        *a.State,
	// 	Country:      *a.Country,
	// 	Latitude:      *a.Latitude,
	// 	Longitude:      *a.Longitude,
	// }

	return address
}
