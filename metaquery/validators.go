package metaquery

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
)

// ValidationResult :: response for Validate
type ValidationResult struct {
	Err       error
	ShouldRun bool
	Message   string
}

type validator func(val string) ValidationResult

// Validate :: validate a full metaquery along with arguments - we can return err & validationResult
func Validate(query string) ValidationResult {
	query = strings.TrimSuffix(query, ";")
	// get the meta query
	q := strings.Split(query, " ")

	validatorFunction := metaQueryDefinitions[q[0]].validator

	if validatorFunction != nil {
		return validatorFunction(strings.Join(getArguments(query), " "))
	}
	return ValidationResult{Err: fmt.Errorf("'%s' is not a known command", query)}
}

func booleanValidator(metaquery string, validators ...validator) validator {
	return func(val string) ValidationResult {
		//	Error: argument required multi-line mode is off.  You can enable it with: .multi on
		//	headers mode is off.  You can enable it with: .headers on
		//	timing mode is off.  You can enable it with: .timing on
		title := metaQueryDefinitions[metaquery].title
		args := strings.Fields(strings.TrimSpace(val))
		numArgs := len(args)

		if numArgs == 0 {
			return ValidationResult{
				Message: fmt.Sprintf(`%s mode is off. You can enable it with: %s on `, title, metaquery),
			}
		}
		if numArgs > 1 {
			return ValidationResult{
				Err: fmt.Errorf("command needs %d argument(s) - got %d", 1, numArgs),
			}
		}
		return buildValidationResult(val, validators)
	}
}

func composeValidator(validators ...validator) validator {
	return func(val string) ValidationResult {
		return buildValidationResult(val, validators)
	}
}

var atMostNArgs = func(n int) validator {
	return func(val string) ValidationResult {
		args := strings.Fields(strings.TrimSpace(val))
		numArgs := len(args)
		if numArgs > n {
			return ValidationResult{
				Err: fmt.Errorf("command needs at most %d argument(s) - got %d", n, numArgs),
			}
		}
		return ValidationResult{ShouldRun: true}
	}
}

var exactlyNArgs = func(n int) validator {
	return func(val string) ValidationResult {
		args := strings.Fields(strings.TrimSpace(val))
		numArgs := len(args)
		if numArgs != n {
			return ValidationResult{
				Err: fmt.Errorf("command needs %d argument(s) - got %d", n, numArgs),
			}
		}
		return ValidationResult{
			ShouldRun: true,
		}
	}
}

var noArgs = exactlyNArgs(0)

var allowedArgValues = func(caseSensitive bool, allowedValues ...string) validator {
	return func(val string) ValidationResult {
		if !caseSensitive {
			// convert everything to lower case
			val = strings.ToLower(val)
			for idx, av := range allowedValues {
				allowedValues[idx] = strings.ToLower(av)
			}
		}
		args := strings.Fields(strings.TrimSpace(val))
		for _, arg := range args {
			if !helpers.StringSliceContains(allowedValues, arg) {
				return ValidationResult{
					Err: fmt.Errorf("valid values for this command are %v - got %s", allowedValues, arg),
				}
			}
		}
		return ValidationResult{ShouldRun: true}
	}
}

func buildValidationResult(val string, validators []validator) ValidationResult {
	var messages string
	for _, v := range validators {
		validate := v(val)
		if validate.Message != "" {
			messages = fmt.Sprintf("%s\n%s", messages, validate.Message)
		}
		if validate.Err != nil {
			return ValidationResult{
				Err:     validate.Err,
				Message: messages,
			}
		}
		if !validate.ShouldRun {
			return ValidationResult{
				Message:   messages,
				ShouldRun: false,
			}
		}
	}
	return ValidationResult{
		Message:   messages,
		ShouldRun: true,
	}
}
