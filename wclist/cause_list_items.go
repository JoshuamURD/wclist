package wclist

//CLIItems represents a cause list item
type CLIItems struct {
	MatterNumber   uint64
	TenementNumber string
	Comments       string
}

// ObjectionItems represents an objection item
type ObjectionItems struct {
	CLIItems
	ObjectionNumber uint64
	ObjectorName    string
	ApplicantName   string
}

// ForfeitureItems represents a forfeiture item
type ForfeitureItems struct {
	CLIItems
	ApplicantName  string
	RespondentName string
}

// ExemptionItems represents an exemption item
type ExemptionItems struct {
	CLIItems
	ApplicantName  string
	RespondentName string
}

// Implement the CauseListItem interface for ObjectionItems
func (c ObjectionItems) GetMatterNumber() uint64    { return c.MatterNumber }
func (c ObjectionItems) GetTenementNumber() string  { return c.TenementNumber }
func (c ObjectionItems) GetComments() string        { return c.Comments }
func (c ObjectionItems) GetApplyingParty() string   { return c.ApplicantName }
func (c ObjectionItems) GetRespondingParty() string { return c.ObjectorName }

// Additional method specific to ObjectionItems
func (c ObjectionItems) GetObjectionNumber() uint64 { return c.ObjectionNumber }

// Implement the CauseListItem interface for ForfeitureItems
func (f ForfeitureItems) GetMatterNumber() uint64    { return f.MatterNumber }
func (f ForfeitureItems) GetTenementNumber() string  { return f.TenementNumber }
func (f ForfeitureItems) GetComments() string        { return f.Comments }
func (f ForfeitureItems) GetApplyingParty() string   { return f.ApplicantName }
func (f ForfeitureItems) GetRespondingParty() string { return f.RespondentName }

// Implement the CauseListItem interface for ExemptionItems
func (e ExemptionItems) GetMatterNumber() uint64    { return e.MatterNumber }
func (e ExemptionItems) GetTenementNumber() string  { return e.TenementNumber }
func (e ExemptionItems) GetComments() string        { return e.Comments }
func (e ExemptionItems) GetApplyingParty() string   { return e.ApplicantName }
func (e ExemptionItems) GetRespondingParty() string { return e.RespondentName }
