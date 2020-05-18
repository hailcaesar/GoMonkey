
package ast
import (
    "gomonkey/token"
    "testing"
)
func TestString(t *testing.T) {
    code := &Code{
        Statements: []Statement{
            &LetStatement{
                Token: token.Token{Type: token.LET, Literal: "let"},
                Name: &Identifier{
                    Token: token.Token{Type: token.IDN, Literal: "RickAndMorty"},
                    Value: "RickAndMorty",

                },
                Value: &Identifier{
                    Token: token.Token{Type: token.IDN, Literal: "BirdMan"},
                    Value: "BirdMan",
                },
            },
        },
    }
    if code.String() != "let RickAndMorty = BirdMan;" {
        t.Errorf("Error: code.String() not working properly [%q]", code.String())
    }
}
