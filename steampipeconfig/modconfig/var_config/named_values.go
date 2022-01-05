package var_config

// github.com/hashicorp/terraform/configs/parser_config.go
import (
	"fmt"
	"unicode"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/steampipeconfig/inputvars/typeexpr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	//"github.com/hashicorp/terraform/internal/typeexpr"
)

// A consistent detail message for all "not a valid identifier" diagnostics.
const badIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."

// Variable represents a "variable" block in a module or file.
type Variable struct {
	Name        string
	Description string
	Default     cty.Value
	Type        cty.Type
	ParsingMode VariableParsingMode
	//Validations []*VariableValidation
	//Sensitive   bool

	DescriptionSet bool
	//SensitiveSet   bool

	DeclRange hcl.Range
}

func DecodeVariableBlock(block *hcl.Block, override bool) (*Variable, hcl.Diagnostics) {
	v := &Variable{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	// Unless we're building an override, we'll set some defaults
	// which we might override with attributes below. We leave these
	// as zero-value in the override case so we can recognize whether
	// or not they are set when we merge.
	if !override {
		v.Type = cty.DynamicPseudoType
		v.ParsingMode = VariableParseLiteral
	}

	content, diags := block.Body.Content(variableBlockSchema)

	if !hclsyntax.ValidIdentifier(v.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid variable name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	//// Don't allow declaration of variables that would conflict with the
	//// reserved attribute and block type names in a "module" block, since
	//// these won't be usable for child modules.
	//for _, attr := range moduleBlockSchema.Attributes {
	//	if attr.Name == v.Name {
	//		diags = append(diags, &hcl.Diagnostic{
	//			Severity: hcl.DiagError,
	//			Summary:  "Invalid variable name",
	//			Detail:   fmt.Sprintf("The variable name %q is reserved due to its special meaning inside module blocks.", attr.Name),
	//			Subject:  &block.LabelRanges[0],
	//		})
	//	}
	//}
	//for _, blockS := range moduleBlockSchema.Blocks {
	//	if blockS.Type == v.Name {
	//		diags = append(diags, &hcl.Diagnostic{
	//			Severity: hcl.DiagError,
	//			Summary:  "Invalid variable name",
	//			Detail:   fmt.Sprintf("The variable name %q is reserved due to its special meaning inside module blocks.", blockS.Type),
	//			Subject:  &block.LabelRanges[0],
	//		})
	//	}
	//}

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &v.Description)
		diags = append(diags, valDiags...)
		v.DescriptionSet = true
	}

	if attr, exists := content.Attributes["type"]; exists {
		ty, parseMode, tyDiags := decodeVariableType(attr.Expr)
		diags = append(diags, tyDiags...)
		v.Type = ty
		v.ParsingMode = parseMode
	}

	//if attr, exists := content.Attributes["sensitive"]; exists {
	//	valDiags := gohcl.DecodeExpression(attr.Expr, nil, &v.Sensitive)
	//	diags = append(diags, valDiags...)
	//	v.SensitiveSet = true
	//}

	if attr, exists := content.Attributes["default"]; exists {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)

		// Convert the default to the expected type so we can catch invalid
		// defaults early and allow later code to assume validity.
		// Note that this depends on us having already processed any "type"
		// attribute above.
		// However, we can't do this if we're in an override file where
		// the type might not be set; we'll catch that during merge.
		if v.Type != cty.NilType {
			var err error
			val, err = convert.Convert(val, v.Type)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid default value for variable",
					Detail:   fmt.Sprintf("This default value is not compatible with the variable's type constraint: %s.", err),
					Subject:  attr.Expr.Range().Ptr(),
				})
				val = cty.DynamicVal
			}
		}

		v.Default = val
	}

	for _, block := range content.Blocks {
		switch block.Type {

		default:
			// The above cases should be exhaustive for all block types
			// defined in variableBlockSchema
			panic(fmt.Sprintf("unhandled block type %q", block.Type))
		}
	}

	return v, diags
}

func decodeVariableType(expr hcl.Expression) (cty.Type, VariableParsingMode, hcl.Diagnostics) {
	if exprIsNativeQuotedString(expr) {
		// If a user provides the pre-0.12 form of variable type argument where
		// the string values "string", "list" and "map" are accepted, we
		// provide an error to point the user towards using the type system
		// correctly has a hint.
		// Only the native syntax ends up in this codepath; we handle the
		// JSON syntax (which is, of course, quoted within the type system)
		// in the normal codepath below.
		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return cty.DynamicPseudoType, VariableParseHCL, diags
		}
		str := val.AsString()
		switch str {
		case "string":
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid quoted type constraints",
				Detail:   "Terraform 0.11 and earlier required type constraints to be given in quotes, but that form is now deprecated and will be removed in a future version of Terraform. Remove the quotes around \"string\".",
				Subject:  expr.Range().Ptr(),
			})
			return cty.DynamicPseudoType, VariableParseLiteral, diags
		case "list":
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid quoted type constraints",
				Detail:   "Terraform 0.11 and earlier required type constraints to be given in quotes, but that form is now deprecated and will be removed in a future version of Terraform. Remove the quotes around \"list\" and write list(string) instead to explicitly indicate that the list elements are strings.",
				Subject:  expr.Range().Ptr(),
			})
			return cty.DynamicPseudoType, VariableParseHCL, diags
		case "map":
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid quoted type constraints",
				Detail:   "Terraform 0.11 and earlier required type constraints to be given in quotes, but that form is now deprecated and will be removed in a future version of Terraform. Remove the quotes around \"map\" and write map(string) instead to explicitly indicate that the map elements are strings.",
				Subject:  expr.Range().Ptr(),
			})
			return cty.DynamicPseudoType, VariableParseHCL, diags
		default:
			return cty.DynamicPseudoType, VariableParseHCL, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Invalid legacy variable type hint",
				Detail:   `To provide a full type expression, remove the surrounding quotes and give the type expression directly.`,
				Subject:  expr.Range().Ptr(),
			}}
		}
	}

	// First we'll deal with some shorthand forms that the HCL-level type
	// expression parser doesn't include. These both emulate pre-0.12 behavior
	// of allowing a list or map of any element type as long as all of the
	// elements are consistent. This is the same as list(any) or map(any).
	switch hcl.ExprAsKeyword(expr) {
	case "list":
		return cty.List(cty.DynamicPseudoType), VariableParseHCL, nil
	case "map":
		return cty.Map(cty.DynamicPseudoType), VariableParseHCL, nil
	}

	ty, diags := typeexpr.TypeConstraint(expr)
	if diags.HasErrors() {
		return cty.DynamicPseudoType, VariableParseHCL, diags
	}

	switch {
	case ty.IsPrimitiveType():
		// Primitive types use literal parsing.
		return ty, VariableParseLiteral, diags
	default:
		// Everything else uses HCL parsing
		return ty, VariableParseHCL, diags
	}
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *Variable) Required() bool {
	return v.Default == cty.NilVal
}

// VariableParsingMode defines how values of a particular variable given by
// text-only mechanisms (command line arguments and environment variables)
// should be parsed to produce the final value.
type VariableParsingMode rune

// VariableParseLiteral is a variable parsing mode that just takes the given
// string directly as a cty.String value.
const VariableParseLiteral VariableParsingMode = 'L'

// VariableParseHCL is a variable parsing mode that attempts to parse the given
// string as an HCL expression and returns the result.
const VariableParseHCL VariableParsingMode = 'H'

// Parse uses the receiving parsing mode to process the given variable value
// string, returning the result along with any diagnostics.
//
// A VariableParsingMode does not know the expected type of the corresponding
// variable, so it's the caller's responsibility to attempt to convert the
// result to the appropriate type and return to the user any diagnostics that
// conversion may produce.
//
// The given name is used to create a synthetic filename in case any diagnostics
// must be generated about the given string value. This should be the name
// of the configuration variable whose value will be populated from the given
// string.
//
// If the returned diagnostics has errors, the returned value may not be
// valid.
func (m VariableParsingMode) Parse(name, value string) (cty.Value, hcl.Diagnostics) {
	switch m {
	case VariableParseLiteral:
		return cty.StringVal(value), nil
	case VariableParseHCL:
		fakeFilename := fmt.Sprintf("<value for var.%s>", name)
		expr, diags := hclsyntax.ParseExpression([]byte(value), fakeFilename, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return cty.DynamicVal, diags
		}
		val, valDiags := expr.Value(nil)
		diags = append(diags, valDiags...)
		return val, diags
	default:
		// Should never happen
		panic(fmt.Errorf("Parse called on invalid VariableParsingMode %#v", m))
	}
}

// VariableValidation represents a configuration-defined validation rule
// for a particular input variable, given as a "validation" block inside
// a "variable" block.
type VariableValidation struct {
	// Condition is an expression that refers to the variable being tested
	// and contains no other references. The expression must return true
	// to indicate that the value is valid or false to indicate that it is
	// invalid. If the expression produces an error, that's considered a bug
	// in the module defining the validation rule, not an error in the caller.
	Condition hcl.Expression

	// ErrorMessage is one or more full sentences, which would need to be in
	// English for consistency with the rest of the error message output but
	// can in practice be in any language as long as it ends with a period.
	// The message should describe what is required for the condition to return
	// true in a way that would make sense to a caller of the module.
	ErrorMessage string

	DeclRange hcl.Range
}

// looksLikeSentence is a simple heuristic that encourages writing error
// messages that will be presentable when included as part of a larger
// Terraform error diagnostic whose other text is written in the Terraform
// UI writing style.
//
// This is intentionally not a very strong validation since we're assuming
// that module authors want to write good messages and might just need a nudge
// about Terraform's specific style, rather than that they are going to try
// to work around these rules to write a lower-quality message.
func looksLikeSentences(s string) bool {
	if len(s) < 1 {
		return false
	}
	runes := []rune(s) // HCL guarantees that all strings are valid UTF-8
	first := runes[0]
	last := runes[len(runes)-1]

	// If the first rune is a letter then it must be an uppercase letter.
	// (This will only see the first rune in a multi-rune combining sequence,
	// but the first rune is generally the letter if any are, and if not then
	// we'll just ignore it because we're primarily expecting English messages
	// right now anyway, for consistency with all of Terraform's other output.)
	if unicode.IsLetter(first) && !unicode.IsUpper(first) {
		return false
	}

	// The string must be at least one full sentence, which implies having
	// sentence-ending punctuation.
	// (This assumes that if a sentence ends with quotes then the period
	// will be outside the quotes, which is consistent with Terraform's UI
	// writing style.)
	return last == '.' || last == '?' || last == '!'
}

var variableBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "description",
		},
		{
			Name: "default",
		},
		{
			Name: "type",
		},
		{
			Name: "sensitive",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "validation",
		},
	},
}

var variableValidationBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "condition",
			Required: true,
		},
		{
			Name:     "error_message",
			Required: true,
		},
	},
}
