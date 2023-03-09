package metaquery

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

// ValidationResult :: response for Validate
type ValidationResult struct {
	Err       error
	ShouldRun bool
	Message   string
}

type validator func(val []string) ValidationResult

// Validate :: validate a full metaquery along with arguments - we can return err & validationResult
func Validate(query string) ValidationResult {
	query = strings.TrimSuffix(query, ";")
	// get the meta query
	cmd, args := getCmdAndArgs(query)

	validatorFunction := metaQueryDefinitions[cmd].validator

	if validatorFunction != nil {
		return validatorFunction(args)
	}
	return ValidationResult{Err: fmt.Errorf("'%s' is not a known command", query)}
}

func booleanValidator(metaquery string, validators ...validator) validator {
	return func(args []string) ValidationResult {
		//	Error: argument required multi-line mode is off.  You can enable it with: .multi on
		//	headers mode is off.  You can enable it with: .headers on
		//	timing mode is off.  You can enable it with: .timing on
		title := metaQueryDefinitions[metaquery].title
		numArgs := len(args)

		if numArgs == 0 {
			// get the current status of this mode (convert metaquery name into arg name)
			// NOTE - request second arg from cast even though we donl;t use it - to avoid panic
			currentStatus := cmdconfig.Viper().GetBool(constants.ArgFromMetaquery(metaquery))
			// what is the new status (the opposite)
			newStatus := !currentStatus

			// convert current and new status to on/off
			currentStatusString := constants.BoolToOnOff(currentStatus)
			newStatusString := constants.BoolToOnOff(newStatus)

			// what is the action to get to the new status
			actionString := constants.BoolToEnableDisable(newStatus)

			return ValidationResult{
				Message: fmt.Sprintf(`%s mode is %s. You can %s it with: %s `,
					title,
					constants.Bold(currentStatusString),
					actionString,
					constants.Bold(fmt.Sprintf("%s %s", metaquery, newStatusString))),
			}
		}
		if numArgs > 1 {
			return ValidationResult{
				Err: fmt.Errorf("command needs 1 argument - got %d", numArgs),
			}
		}
		return buildValidationResult(args, validators)
	}
}

func composeValidator(validators ...validator) validator {
	return func(val []string) ValidationResult {
		return buildValidationResult(val, validators)
	}
}

func validatorFromArgsOf(cmd string) validator {
	return func(val []string) ValidationResult {
		metaQueryDefinition := metaQueryDefinitions[cmd]
		validArgs := []string{}

		for _, validArg := range metaQueryDefinition.args {
			validArgs = append(validArgs, validArg.value)
		}

		return allowedArgValues(false, validArgs...)(val)
	}
}

var atLeastNArgs = func(n int) validator {
	return func(args []string) ValidationResult {
		numArgs := len(args)
		if numArgs < n {
			return ValidationResult{
				Err: fmt.Errorf("command needs at least %d %s - got %d", n, utils.Pluralize("argument", n), numArgs),
			}
		}
		return ValidationResult{ShouldRun: true}
	}
}

var atMostNArgs = func(n int) validator {
	return func(args []string) ValidationResult {
		numArgs := len(args)
		if numArgs > n {
			return ValidationResult{
				Err: fmt.Errorf("command needs at most %d %s - got %d", n, utils.Pluralize("argument", n), numArgs),
			}
		}
		return ValidationResult{ShouldRun: true}
	}
}

var exactlyNArgs = func(n int) validator {
	return func(args []string) ValidationResult {
		numArgs := len(args)
		if numArgs != n {
			return ValidationResult{
				Err: fmt.Errorf("command needs %d %s - got %d", n, utils.Pluralize("argument", n), numArgs),
			}
		}
		return ValidationResult{
			ShouldRun: true,
		}
	}
}

var noArgs = exactlyNArgs(0)

var allowedArgValues = func(caseSensitive bool, allowedValues ...string) validator {
	return func(args []string) ValidationResult {
		if !caseSensitive {
			// convert everything to lower case
			for idx, a := range args {
				args[idx] = strings.ToLower(a)
			}
			for idx, av := range allowedValues {
				allowedValues[idx] = strings.ToLower(av)
			}
		}

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

func buildValidationResult(val []string, validators []validator) ValidationResult {
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
