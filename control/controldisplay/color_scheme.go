package controldisplay

import (
	"fmt"
	"reflect"

	"github.com/logrusorgru/aurora"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

type colorFunc func(interface{}) aurora.Value

// ControlColors is a global variable containing the current control color scheme
var ControlColors *ControlColorScheme

type ControlColorSchemaDefinition struct {
	// group
	GroupTitle           string
	Severity             string
	CountZeroFail        string
	CountZeroFailDivider string
	CountDivider         string
	CountFail            string
	CountTotal           string
	CountTotalAllPassed  string
	CountGraphFail       string
	CountGraphPass       string
	CountGraphBracket    string

	// results
	StatusAlarm string
	StatusError string
	StatusSkip  string
	StatusInfo  string
	StatusOK    string
	StatusColon string
	ReasonAlarm string
	ReasonError string
	ReasonSkip  string
	ReasonInfo  string
	ReasonOK    string

	Spacer   string
	UseColor bool
}

type ControlColorScheme struct {
	GroupTitle           colorFunc
	Severity             colorFunc
	CountZeroFail        colorFunc
	CountZeroFailDivider colorFunc
	CountDivider         colorFunc
	CountFail            colorFunc
	CountTotal           colorFunc
	CountTotalAllPassed  colorFunc
	CountGraphFail       colorFunc
	CountGraphPass       colorFunc
	CountGraphBracket    colorFunc
	StatusAlarm          colorFunc
	StatusError          colorFunc
	StatusSkip           colorFunc
	StatusInfo           colorFunc
	StatusOK             colorFunc
	StatusColon          colorFunc
	ReasonAlarm          colorFunc
	ReasonError          colorFunc
	ReasonSkip           colorFunc
	ReasonInfo           colorFunc
	ReasonOK             colorFunc
	Spacer               colorFunc

	ReasonColors map[string]colorFunc
	StatusColors map[string]colorFunc
	UseColor     bool
}

func NewControlColorScheme(def *ControlColorSchemaDefinition) (*ControlColorScheme, error) {
	res := &ControlColorScheme{
		ReasonColors: make(map[string]colorFunc),
		StatusColors: make(map[string]colorFunc),
	}
	err := res.Initialise(def)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *ControlColorScheme) Initialise(def *ControlColorSchemaDefinition) error {
	destV := reflect.ValueOf(c).Elem()

	nullColorFunc := func(val interface{}) aurora.Value { return aurora.Reset(val) }
	var validationErrors []string

	v := reflect.ValueOf(def).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)

		// all string fields are colors skip non string fields
		if fieldType.Type.Name() != "string" {
			continue
		}

		colorString := fieldValue.Interface().(string)
		property := fieldType.Name
		// find corresponding field in dest
		destField := destV.FieldByName(property)

		// if no color is set, use null color function
		if colorString == "" {
			destField.Set(reflect.ValueOf(nullColorFunc))
			continue
		}

		// is this a valid color string?
		if f, ok := constants.Colors[colorString]; ok {
			destField.Set(reflect.ValueOf(f))
		} else {
			validationErrors = append(validationErrors, property)
		}
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf("invalid color scheme. %d %s have invalid colors: %v",
			len(validationErrors),
			utils.Pluralize("property", len(validationErrors)),
			validationErrors)
	}
	// populate the color maps
	c.ReasonColors["alarm"] = c.ReasonAlarm
	c.ReasonColors["skip"] = c.ReasonSkip
	c.ReasonColors["info"] = c.ReasonInfo
	c.ReasonColors["error"] = c.ReasonError
	c.ReasonColors["ok"] = c.ReasonOK
	c.StatusColors["alarm"] = c.StatusAlarm
	c.StatusColors["skip"] = c.StatusSkip
	c.StatusColors["info"] = c.StatusInfo
	c.StatusColors["error"] = c.StatusError
	c.StatusColors["ok"] = c.StatusOK
	c.UseColor = def.UseColor
	return nil
}

func (c ControlColorScheme) initialiseColor(color string, dest *colorFunc, validationErrors []string) {
	if f, ok := constants.Colors[color]; ok {
		*dest = f
	} else {
		validationErrors = append(validationErrors, color)
	}
}

var ColorSchemes = map[string]*ControlColorSchemaDefinition{
	"dark": {
		GroupTitle:           "bold-bright-white",
		Severity:             "bold-bright-yellow",
		CountZeroFail:        "gray1",
		CountZeroFailDivider: "gray1",
		CountDivider:         "gray2",
		CountFail:            "bold-bright-red",
		CountTotal:           "bright-white",
		CountTotalAllPassed:  "bold-bright-green",
		CountGraphFail:       "bright-red",
		CountGraphPass:       "bright-green",
		CountGraphBracket:    "gray2",
		StatusAlarm:          "bold-bright-red",
		StatusError:          "bold-bright-red",
		StatusSkip:           "gray3",
		StatusInfo:           "bright-cyan",
		StatusOK:             "bright-green",
		StatusColon:          "gray1",
		ReasonAlarm:          "bright-red",
		ReasonError:          "bright-red",
		ReasonSkip:           "gray3",
		ReasonInfo:           "bright-cyan",
		ReasonOK:             "gray4",
		Spacer:               "gray1",
		UseColor:             true,
	},
	"light": {
		GroupTitle:           "bold-bright-black",
		Severity:             "bold-bright-yellow",
		CountZeroFail:        "gray5",
		CountZeroFailDivider: "gray5",
		CountDivider:         "gray4",
		CountFail:            "bold-bright-red",
		CountTotal:           "bright-black",
		CountTotalAllPassed:  "bold-bright-green",
		CountGraphFail:       "bright-red",
		CountGraphPass:       "bright-green",
		CountGraphBracket:    "gray4",
		StatusAlarm:          "bold-bright-red",
		StatusError:          "bold-bright-red",
		StatusSkip:           "gray3",
		StatusInfo:           "bright-cyan",
		StatusOK:             "bright-green",
		StatusColon:          "gray5",
		ReasonAlarm:          "bright-red",
		ReasonError:          "bright-red",
		ReasonSkip:           "gray3",
		ReasonInfo:           "bright-cyan",
		ReasonOK:             "gray2",
		Spacer:               "gray5",
		UseColor:             true,
	},
	"plain": {UseColor: false},
}
