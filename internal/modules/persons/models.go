package persons

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var ErrNotFound = errors.New("person not found")

type Person struct {
	ID                  string    `json:"id"`
	FullName            string    `json:"full_name"`
	PhoneNumber         *string   `json:"phone_number,omitempty"`
	Email               *string   `json:"email,omitempty"`
	BirthDate           *string   `json:"birth_date,omitempty"` // YYYY-MM-DD
	RegistrationAddress *string   `json:"registration_address,omitempty"`
	FactAddress         *string   `json:"fact_address,omitempty"`
	IsDeleted           bool      `json:"is_deleted"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type CreateInput struct {
	FullName            string  `json:"full_name"`
	PhoneNumber         *string `json:"phone_number,omitempty"`
	Email               *string `json:"email,omitempty"`
	BirthDate           *string `json:"birth_date,omitempty"`
	RegistrationAddress *string `json:"registration_address,omitempty"`
	FactAddress         *string `json:"fact_address,omitempty"`
}

type UpdateInput struct {
	FullName            *string `json:"full_name,omitempty"`
	PhoneNumber         *string `json:"phone_number,omitempty"`
	Email               *string `json:"email,omitempty"`
	BirthDate           *string `json:"birth_date,omitempty"`
	RegistrationAddress *string `json:"registration_address,omitempty"`
	FactAddress         *string `json:"fact_address,omitempty"`
}

type ListFilter struct {
	Search string
	Limit  int
	Offset int
}

var (
	reDate  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	rePhone = regexp.MustCompile(`[^\d+]`)
)

func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	s = rePhone.ReplaceAllString(s, "")
	if strings.HasPrefix(s, "+") {
		return "+" + strings.ReplaceAll(s[1:], "+", "")
	}
	return strings.ReplaceAll(s, "+", "")
}

func trimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func (in *CreateInput) Normalize() {
	in.FullName = strings.TrimSpace(in.FullName)

	if in.PhoneNumber != nil {
		v := normalizePhone(*in.PhoneNumber)
		if v == "" {
			in.PhoneNumber = nil
		} else {
			in.PhoneNumber = &v
		}
	}
	if in.Email != nil {
		v := strings.ToLower(strings.TrimSpace(*in.Email))
		if v == "" {
			in.Email = nil
		} else {
			in.Email = &v
		}
	}
	in.BirthDate = trimPtr(in.BirthDate)
	in.RegistrationAddress = trimPtr(in.RegistrationAddress)
	in.FactAddress = trimPtr(in.FactAddress)
}

func (in CreateInput) Validate() error {
	if strings.TrimSpace(in.FullName) == "" {
		return errors.New("full_name is required")
	}
	if in.BirthDate != nil && !reDate.MatchString(*in.BirthDate) {
		return errors.New("birth_date must be YYYY-MM-DD")
	}
	return nil
}

func (in *UpdateInput) Normalize() {
	if in.FullName != nil {
		v := strings.TrimSpace(*in.FullName)
		in.FullName = &v
	}
	if in.PhoneNumber != nil {
		v := normalizePhone(*in.PhoneNumber)
		if v == "" {
			empty := ""
			in.PhoneNumber = &empty
		} else {
			in.PhoneNumber = &v
		}
	}
	if in.Email != nil {
		v := strings.ToLower(strings.TrimSpace(*in.Email))
		in.Email = &v
	}
	if in.BirthDate != nil {
		v := strings.TrimSpace(*in.BirthDate)
		in.BirthDate = &v
	}
	if in.RegistrationAddress != nil {
		v := strings.TrimSpace(*in.RegistrationAddress)
		in.RegistrationAddress = &v
	}
	if in.FactAddress != nil {
		v := strings.TrimSpace(*in.FactAddress)
		in.FactAddress = &v
	}
}

func (in UpdateInput) Validate() error {
	if in.FullName != nil && strings.TrimSpace(*in.FullName) == "" {
		return errors.New("full_name cannot be empty")
	}
	if in.BirthDate != nil && *in.BirthDate != "" && !reDate.MatchString(*in.BirthDate) {
		return errors.New("birth_date must be YYYY-MM-DD")
	}
	return nil
}
