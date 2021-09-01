package parse

import (
	"fmt"
	"testing"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type parsePreparedStatementInvocationTest struct {
	input    string
	expected parsePreparedStatementInvocationResult
}

type parsePreparedStatementInvocationResult struct {
	queryName string
	params    *modconfig.QueryParams
}

var emptyParams = modconfig.NewQueryParams()
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
			params:    &modconfig.QueryParams{ParamsList: []string{"foo"}},
		},
	},
	"invalid params 2": {
		input: `query.q1("foo")`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{ParamsList: []string{`"foo"`}},
		},
	},
	"invalid params 3": {
		input: `query.q1('foo',  'bar')`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{ParamsList: []string{"'foo'", "'bar'"}},
		},
	},
	"invalid params 4": {
		input: `query.q1(['foo',  'bar'])`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{ParamsList: []string{"['foo'", "'bar']"}},
		},
	},

	"single positional param": {
		input: `query.q1('foo')`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{ParamsList: []string{"'foo'"}},
		},
	},
	"single positional param extra spaces": {
		input: `query.q1('foo')`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{ParamsList: []string{"'foo'"}},
		},
	},
	"multiple positional params": {
		input: `query.q1('foo', 'bar', 'foo-bar')`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{ParamsList: []string{"'foo'", "'bar'", "'foo-bar'"}},
		},
	},
	"multiple positional params extra spaces": {
		input: ` query.q1('foo' ,  'bar', 'foo-bar'  )`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{ParamsList: []string{"'foo'", "'bar'", "'foo-bar'"}},
		},
	},
	"single named param": {
		input: `query.q1(p1 => 'foo')`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{Params: map[string]string{"p1": "foo"}},
		},
	},
	"single named param extra spaces": {
		input: `query.q1( p1  =>  'foo' ) `,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{Params: map[string]string{"p1": "foo"}},
		},
	},
	"multiple named params": {
		input: `query.q1(p1 => 'foo', p2 => 'bar')`,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{Params: map[string]string{"p1": "foo", "p2": "bar"}},
		},
	},
	"multiple named params extra spaces": {
		input: ` query.q1 ( p1 => 'foo' ,  p2  => 'bar'     ) `,
		expected: parsePreparedStatementInvocationResult{
			queryName: `query.q1`,
			params:    &modconfig.QueryParams{Params: map[string]string{"p1": "foo", "p2": "bar"}},
		},
	},
}

func TestParsePreparedStatementInvocation(t *testing.T) {
	for name, test := range testCasesParsePreparedStatementInvocation {
		queryName, params := ParsePreparedStatementInvocation(test.input)

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
