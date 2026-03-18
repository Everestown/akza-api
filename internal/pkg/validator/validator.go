package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	reSlug        = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	reTgUsername  = regexp.MustCompile(`^[a-zA-Z0-9_]{5,32}$`)
)

// RegisterCustomValidators adds project-specific validation tags.
func RegisterCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return reSlug.MatchString(fl.Field().String())
	})
	_ = v.RegisterValidation("tg_username", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		// Strip leading @ if present
		if len(s) > 0 && s[0] == '@' {
			s = s[1:]
		}
		return reTgUsername.MatchString(s)
	})
}
