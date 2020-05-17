package token

type TokenType string

type Token struct {
    Type    TokenType
    Literal string
}

var keywords = map[string]TokenType{
    "if"     : IF,
    "else"   : ELSE,
    "return" : RET,
    "fn"     : FNCT,
    "let"    : LET,
    "true"   : TRUE,
    "false"  : FALS,
}

func IdentifierLookup(identifier string) TokenType{
    if tok, ok := keywords[identifier]; ok {
        return tok
    }
    return IDN
}

const (
    //Keywords
    LET     = "let"
    FNCT    = "function"
    TRUE    = "true"
    FALS    = "false"
    IF      = "if"
    ELSE    = "else"
    RET     = "ret"
    EQ      = "=="
    NEQ     = "!="

    //Identifiers + Literals
    IDN     = "identifier"
    INT     = "int"

    //Delimeters
    COM     = ","
    SCLN    = ";"

    LPAR    = "("
    RPAR    = ")"
    LBRA    = "{"
    RBRA    = "}"
    
    //Operators
    AGMT    = "="
    PLUS    = "+"
    DIV     = "/"
    MINS    = "-"
    BANG    = "!"
    ASTK    = "*"
    LT      = "<"
    GT      = ">"

    
    //Misc
    ERR     = "illegal"
    EOF     = "eof"
)
