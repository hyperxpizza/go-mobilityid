package din

import (
	"fmt"
	"regexp"
	"strings"

	c "github.com/hyperxpizza/go-mobilityid/common"
	"github.com/hyperxpizza/go-mobilityid/contractid"
)

const instanceMaxLength = 6

var regex = regexp.MustCompile(fmt.Sprintf("^(?P<country>%v)(?:[*-]?)(?P<party>%v)(?:[*-]?)(?P<instance>%v)(?:(?:[*-]?)(?P<check>%v))?$", c.CountryCodeRegex, c.PartyCodeRegex, "([A-Za-z0-9]{6})", c.CheckDigitRegex))

type ContractId struct {
	contractid.Reader
}

// NewContractIdNoCheckDigit returns a DIN contract ID complete of check digit, if provided input is valid; returns an error otherwise.
func NewContractIdNoCheckDigit(countryCode, partyCode, instance string) (*ContractId, error) {
	if err := contractid.ValidateNoCheckDigit(countryCode, partyCode, instance, instanceMaxLength); err != nil {
		return nil, err
	}

	return &ContractId{
		contractid.Id(
			strings.ToUpper(countryCode),
			strings.ToUpper(partyCode),
			strings.ToUpper(instance),
			ComputeCheckDigit(countryCode+partyCode+instance),
		),
	}, nil
}

// NewContractId returns a DIN contract ID, if provided input is valid; returns an error otherwise.
func NewContractId(countryCode, partyCode, instance string, checkDigit rune) (*ContractId, error) {
	id, err := NewContractIdNoCheckDigit(countryCode, partyCode, instance)
	if err != nil {
		return nil, err
	}

	if checkDigit != id.CheckDigit() {
		return nil, fmt.Errorf("provided check digit '%c' doesn't match computed one '%c'", checkDigit, id.CheckDigit())
	}

	return id, nil
}

// Parse parses the input string into a DIN contract ID, if it is valid; returns an error otherwise.
// A check digit will only be present, in returned struct, if the provided string contained it.
func Parse(input string) (*ContractId, error) {
	groups := regex.FindStringSubmatch(input)

	countryCode, err := c.ExtractAndUpcaseGroup(regex, groups, "country", true)
	if err != nil {
		return nil, fmt.Errorf("not a DIN contract ID: %v", input)
	}
	partyCode, err := c.ExtractAndUpcaseGroup(regex, groups, "party", true)
	if err != nil {
		return nil, fmt.Errorf("not a DIN contract ID: %v", input)
	}
	instance, err := c.ExtractAndUpcaseGroup(regex, groups, "instance", true)
	if err != nil {
		return nil, fmt.Errorf("not a DIN contract ID: %v", input)
	}
	check, err := c.ExtractAndUpcaseGroup(regex, groups, "check", false)
	if err != nil {
		return nil, fmt.Errorf("not a DIN contract ID: %v", input)
	}

	var checkDigit rune
	if len(check) > 0 {
		checkDigit = rune(check[0])
		if err := validate(countryCode, partyCode, instance, checkDigit); err != nil {
			return nil, err
		}
	} else if err := contractid.ValidateNoCheckDigit(countryCode, partyCode, instance, instanceMaxLength); err != nil {
		return nil, err
	}

	return &ContractId{
		contractid.Id(
			countryCode,
			partyCode,
			instance,
			ComputeCheckDigit(countryCode+partyCode+instance),
		),
	}, nil
}

func validate(countryCode, partyCode, instance string, checkDigit rune) error {
	if err := contractid.ValidateNoCheckDigit(countryCode, partyCode, instance, instanceMaxLength); err != nil {
		return err
	}

	if checkDigit != ComputeCheckDigit(countryCode+partyCode+instance) {
		return fmt.Errorf("check digit '%c' is invalid", checkDigit)
	}

	return nil
}
