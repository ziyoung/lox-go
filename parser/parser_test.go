package parser

import (
	"testing"

	"github.com/ziyoung/lox-go/ast"
	"github.com/ziyoung/lox-go/lexer"
)

type parserTest struct {
	input    string
	expected string
}

func TestParseExpression(t *testing.T) {
	tests := []parserTest{
		{
			input:    "123 - -456 - 789",
			expected: "((123 - (-456)) - 789)",
		},
		{
			input:    "1 >= 2 == !false",
			expected: "((1 >= 2) == (!false))",
		},
		{
			input:    "123 - 456 * 789 / 123",
			expected: "(123 - ((456 * 789) / 123))",
		},
	}
	for i, test := range tests {
		p := newParserFromInput(test.input)
		expr, err := parseExpression(p)
		if err != nil {
			t.Fatalf("test [%d]: parse failed. error is %s", i, err.Error())
		}
		if expr.String() != test.expected {
			t.Fatalf("test [%d]: expected expression is %q. got %q.", i, test.expected, expr.String())
		}
	}
}

func TestParseLogicExpr(t *testing.T) {
	tests := []parserTest{
		{
			input:    "true or false",
			expected: "true or false",
		},
		{
			input:    "true and false",
			expected: "true and false",
		},
		{
			input:    "false or (a = 2)",
			expected: "false or (a = 2)",
		},
		{
			input:    "a = 2 and false",
			expected: "a = 2 and false",
		},
	}
	for i, test := range tests {
		p := newParserFromInput(test.input)
		expr, err := parseExpression(p)
		if err != nil {
			t.Fatalf("test [%d] parse failed. error is %s", i, err.Error())
		}
		if expr.String() != test.expected {
			t.Fatalf("test [%d] expected expression is %q. got %q.", i, test.expected, expr.String())
		}
	}
}

func TestParseExpressionRecover(t *testing.T) {
	input := "123 + 456 -;123+456"
	expected := "(123 + 456)"
	p := newParserFromInput(input)
	expr, err := parseExpression(p)
	if err == nil {
		t.Fatalf("parser doesn't fail. get expression %s", expr)
	}

	p.synchronize()
	expr, err = parseExpression(p)
	if err != nil {
		t.Fatalf("error occurs. error is %s", err.Error())
	}
	if expr.String() != expected {
		t.Fatalf("expected expression is %q. got %q.", expected, expr.String())
	}
}

func TestParsePrintStatement(t *testing.T) {
	input := `var a = 0;
		a = a + 10;
		var b = a = a + 100;
		print a;`
	expected := []string{
		"var a = 0;",
		"a = (a + 10);",
		"var b = a = (a + 100);",
		"print a;",
	}
	p := newParserFromInput(input)
	statements, err := p.Parse()
	if err != nil {
		t.Fatalf("parse failed. error: %s", err.Error())
	}
	if len(statements) != len(expected) {
		t.Fatalf("length of statements should be %d. got %d", len(expected), len(statements))
	}
	for i, stmt := range statements {
		if stmt.String() != expected[i] {
			t.Errorf("test [%d]: expected text is %q. got %q", i, expected[i], stmt.String())
		}
	}
}

func TestParseConditionStatement(t *testing.T) {
	increment := "a = (a + 1);"
	printStmt := "print a;"
	block := func(s string) string {
		return "{ " + s + " }"
	}
	whileStmt := func(s string) string {
		body := printStmt
		if s != "" {
			body = block(body) + s
		}
		return "while ((a < 2)) " + block(body)
	}

	tests := []parserTest{
		{
			input: `if (a) {
				print a + 1;
			} else if (a > 1) {
				print a + 1;
			}`,
			expected: `if (a) { print (a + 1); } else if ((a > 1)) { print (a + 1); }`,
		},
		{
			input: `if (a)
				print a + 1;`,
			expected: `if (a) print (a + 1);`,
		},
		{
			input: `while (a)
				print a;`,
			expected: "while (a) print a;",
		},
		{
			input: `while (a < 2) {
				print a;
			}`,
			expected: whileStmt(""),
		},
		{
			input: `for (var a = 1; a < 2; a = a + 1) {
				print a;
			}`,
			expected: block("var a = 1;" + whileStmt(increment)),
		},

		{
			input: `for (a = 1; a < 2; a = a + 1) {
				print a;
			}`,
			expected: block("a = 1;" + whileStmt(increment)),
		},
		{
			input: `for (; a < 2; a = a + 1) {
				print a;
			}`,
			expected: whileStmt(increment),
		},
		{
			input: `for (;;a = a + 1) {
				print a;
			}`,
			expected: "while (true) " + block(block(printStmt)+increment),
		},
		{
			input: `for (;;) {
				print a;
			}`,
			expected: "while (true) " + block(printStmt),
		},
	}
	for i, test := range tests {
		p := newParserFromInput(test.input)
		statements, err := p.Parse()
		if err != nil {
			t.Fatalf("test [%d]: parse failed. error: %s", i, err.Error())
		}
		if len(statements) != 1 {
			t.Fatalf("test [%d]: should have 1 statement. got %d", i, len(statements))
		}
		s := statements[0].String()
		if s != test.expected {
			t.Fatalf("test [%d]: expected content is %q. got %q", i, test.expected, s)
		}
	}
}

func newParserFromInput(input string) *Parser {
	l := lexer.New(input)
	return New(l)
}

func parseExpression(p *Parser) (expr ast.Expr, err error) {
	defer func() {
		if r := recover(); r != nil {
			if parseErr, ok := r.(parseError); ok {
				err = &parseErr
				expr = nil
			} else {
				panic(r)
			}
		}
	}()
	expr = p.parseExpression()
	return expr, nil
}
