package models

type Lawyer struct {
	Name     string
	Email    string
	Phone    string
	Assigned []AssignedMatter
}

type AssignedMatter struct {
	ClientName      string
	TenementNumber  string
	OtherPartyNames []string
}

func NewLawyer(name, email, phone string) *Lawyer {
	return &Lawyer{
		Name:     name,
		Email:    email,
		Phone:    phone,
		Assigned: []AssignedMatter{},
	}
}

func (l *Lawyer) AddAssignedMatter(clientName, tenementNumber string, otherPartyNames []string) {
	l.Assigned = append(l.Assigned, AssignedMatter{
		ClientName:      clientName,
		TenementNumber:  tenementNumber,
		OtherPartyNames: otherPartyNames,
	})
}
