package validator

import (
	"github.com/Masterminds/semver/v3"
	"github.com/go-playground/validator/v10"
)

func solc_version(fl validator.FieldLevel) bool {
	vStr := fl.Field().String()
	version, err := semver.NewVersion(vStr)
	if err != nil {
		return false
	}
	constraint, _ := semver.NewConstraint(">=0.5.0, <=0.8.30") // According to the Solidity version range
	return constraint.Check(version)
}
