package domain

type LeadGrupoOLX struct {
	LeadOrigin      string                `json:"leadOrigin"`
	Timestamp       string                `json:"timestamp"`
	OriginLeadID    string                `json:"originLeadId"`
	OriginListingID string                `json:"originListingId,omitempty"`
	ClientListingID string                `json:"clientListingId,omitempty"`
	Name            string                `json:"name"`
	Email           string                `json:"email,omitempty"`
	DDD             string                `json:"ddd,omitempty"`
	Phone           string                `json:"phone,omitempty"`
	PhoneNumber     string                `json:"phoneNumber,omitempty"`
	Message         string                `json:"message,omitempty"`
	Temperature     string                `json:"temperature,omitempty"`
	TransactionType string                `json:"transactionType,omitempty"`
	ExtraData       ExtraDataLeadGrupoOLX `json:"extraData,omitempty"`
}

type ExtraDataLeadGrupoOLX struct {
	LeadCerto bool   `json:"leadCerto,omitempty"`
	LeadType  string `json:"leadType,omitempty"`
}
