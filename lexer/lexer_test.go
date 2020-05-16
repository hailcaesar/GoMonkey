package lexer

import (
    "testing"
    "gomonkey/token"
)

func TestNextToken(t *testing.T){
    inputs := [...]string{`=+(){},;let`,
                        `let five = 5;
                        let ten = 10;

                        let add = fn(x, y) {
                          x + y;
                        };

                        let result = add(five, ten);
                        !-/*5;
                        5 < 10 > 5;

                        if (5 < 10) {
                            return true;
                        } else {
                            return false;
                        }

                        10 == 10;
                        10 != 9;
                        `}
    type Expected struct{
        expectedType    token.TokenType
        expectedLiteral string
    }


/*
    tests := []Expected{
        {token.AGMT, "="},
        {token.PLUS, "+"},
        {token.LPAR, "("},
        {token.RPAR, ")"},
        {token.LBRA, "{"},
        {token.RBRA, "}"},
        {token.COM, ","},
        {token.SCLN, ";"},
        {token.LET, "let"},
        {token.EOF, "eof"},
    }
    */

    tests := []Expected{
        {token.LET, "let"},
        {token.IDN, "five"},
        {token.AGMT, "="},
        {token.INT, "5"},
        {token.SCLN, ";"},
        {token.LET, "let"},
        {token.IDN, "ten"},
        {token.AGMT, "="},
        {token.INT, "10"},
        {token.SCLN, ";"},
        {token.LET, "let"},
        {token.IDN, "add"},
        {token.AGMT, "="},
        {token.FNCT, "fn"},
        {token.LPAR, "("},
        {token.IDN, "x"},
        {token.COM, ","},
        {token.IDN, "y"},
        {token.RPAR, ")"},
        {token.LBRA, "{"},
        {token.IDN, "x"},
        {token.PLUS, "+"},
        {token.IDN, "y"},
        {token.SCLN, ";"},
        {token.RBRA, "}"},
        {token.SCLN, ";"},
        {token.LET, "let"},
        {token.IDN, "result"},
        {token.AGMT, "="},
        {token.IDN, "add"},
        {token.LPAR, "("},
        {token.IDN, "five"},
        {token.COM, ","},
        {token.IDN, "ten"},
        {token.RPAR, ")"},
        {token.SCLN, ";"},
        {token.BANG, "!"},
        {token.MINS, "-"},
        {token.DIV, "/"},
        {token.ASTK, "*"},
        {token.INT, "5"},
        {token.SCLN, ";"},
        {token.INT, "5"},
        {token.LT, "<"},
        {token.INT, "10"},
        {token.GT, ">"},
        {token.INT, "5"},
        {token.SCLN, ";"},
        {token.IF, "if"},
        {token.LPAR, "("},
        {token.INT, "5"},
        {token.LT, "<"},
        {token.INT, "10"},
        {token.RPAR, ")"},
        {token.LBRA, "{"},
        {token.RET, "return"},
        {token.TRUE, "true"},
        {token.SCLN, ";"},
        {token.RBRA, "}"},
        {token.ELSE, "else"},
        {token.LBRA, "{"},
        {token.RET, "return"},
        {token.FALS, "false"},
        {token.SCLN, ";"},
        {token.RBRA, "}"},
        {token.INT, "10"},
        {token.EQ, "=="},
        {token.INT, "10"},
        {token.SCLN, ";"},
        {token.INT, "10"},
        {token.NEQ, "!="},
        {token.INT, "9"},
        {token.SCLN, ";"},
        {token.EOF, ""},
    }
    
    l := NewLexer(inputs[1])

    for i, tt := range tests {
        tok := l.NextToken()

        if tok.Type != tt.expectedType {
            t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
                i, tt.expectedType, tok.Type)
        }

        if tok.Literal != tt.expectedLiteral {
            t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
                i, tt.expectedLiteral, tok.Literal)
        }
    }
}
