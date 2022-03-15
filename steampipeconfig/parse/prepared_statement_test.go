package parse

import (
	"fmt"
	"testing"

	"github.com/turbot/steampipe/utils"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type parsePreparedStatementInvocationTest struct {
	input    string
	expected parsePreparedStatementInvocationResult
}

type parsePreparedStatementInvocationResult struct {
	queryName string
	params    *modconfig.QueryArgs
}

var emptyParams = modconfig.NewQueryArgs()
var testCasesParsePreparedStatementInvocation = map[string]parsePreparedStatementInvocationTest{
	"no brackets": {
		input:    `query.q1`,
		expected: parsePreparedStatementInvocationResult{"query.q1", emptyParams},
	},
	"no params": {
		input:    `query.q1()`,
		expected: parsePreparedStatementInvocationResult{"query.q1", emptyParams},
	},
	"invalid params 1": {
		input: `query.q1(foo)`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{},
		},
	},
	"invalid params 4": {
		input: `query.q1("foo",  "bar"])`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{},
		},
	},

	"single positional param": {
		input: `query.q1("foo")`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{ArgList: []*string{utils.ToStringPointer("'foo'")}},
		},
	},
	"single positional param extra spaces": {
		input: `query.q1("foo"   )   `,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{ArgList: []*string{utils.ToStringPointer("'foo'")}},
		},
	},
	"multiple positional params": {
		input: `query.q1("foo", "bar", "foo-bar")`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{ArgList: []*string{utils.ToStringPointer("'foo'"), utils.ToStringPointer("'bar'"), utils.ToStringPointer("'foo-bar'")}},
		},
	},
	"multiple positional params extra spaces": {
		input: `query.q1("foo",   "bar",    "foo-bar"   )`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{ArgList: []*string{utils.ToStringPointer("'foo'"), utils.ToStringPointer("'bar'"), utils.ToStringPointer("'foo-bar'")}},
		},
	},
	"single named param": {
		input: `query.q1(p1 => "foo")`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{ArgMap: map[string]string{"p1": "'foo'"}},
		},
	},
	"single named param extra spaces": {
		input: `query.q1(  p1  =>  "foo"  ) `,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{ArgMap: map[string]string{"p1": "'foo'"}},
		},
	},
	"multiple named params": {
		input: `query.q1(p1 => "foo", p2 => "bar")`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{ArgMap: map[string]string{"p1": "'foo'", "p2": "'bar'"}},
		},
	},
	"multiple named params extra spaces": {
		input: ` query.q1 ( p1 => "foo" ,  p2  => "bar"     ) `,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryArgs{ArgMap: map[string]string{"p1": "'foo'", "p2": "'bar'"}},
		},
	},
}

func TestParsePreparedStatementInvocation(t *testing.T) {
	for name, test := range testCasesParsePreparedStatementInvocation {
		queryName, params, _ := ParsePreparedStatementInvocation(test.input)

		if queryName != test.expected.queryName || !test.expected.params.Equals(params) {
			fmt.Printf("")
			t.Errorf("Test: '%s'' FAILED : expected:\nquery: %s params: %s\n\ngot:\nquery: %s params: %s",
				name,
				test.expected.queryName,
				test.expected.params,
				queryName, params)
		}
	}
}
