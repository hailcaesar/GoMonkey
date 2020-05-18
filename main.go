package main

import (
	"fmt"
	"gomonkey/lexer"
	"gomonkey/parser"
)

func main() {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"a * b + c",
			"((a * b) + c)",
		},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		code := p.ParseCode()
		actual := code.String()
		if actual != tt.expected {
			fmt.Println("expected=%q, got=%q", tt.expected, actual)
		}
	}
}
