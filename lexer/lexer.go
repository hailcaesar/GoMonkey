package lexer

import (
 //   "fmt"
	"gomonkey/token"
)

type Lexer struct{
    input       string
    position    int
    readPtr     int
    curChar     byte
}


func  New(code string) *Lexer{
    l := &Lexer{input : code}
    l.readChar()
    return l
}

func (l *Lexer) readChar(){
    if l.readPtr >= len(l.input){
        l.curChar = 0   //ASCII = NUL
    } else {
        l.curChar = l.input[l.readPtr]
    }
    l.position = l.readPtr
    l.readPtr += 1
}


func (l *Lexer) NextToken() token.Token {
    var tok token.Token

    l.skipWhitespace()

    switch l.curChar {
    case '=':
        if l.peek() == '=' {
            tok = token.Token{Type: token.EQ, 
                  Literal: string(l.curChar) + string(l.input[l.readPtr])}
            l.readChar();
        } else {
            tok = newToken(token.AGMT, l.curChar)
        }
    case ';':
        tok = newToken(token.SCLN, l.curChar)
    case '(':
        tok = newToken(token.LPAR, l.curChar)
    case ')':
        tok = newToken(token.RPAR, l.curChar)
    case ',':
        tok = newToken(token.COM, l.curChar)
    case '+':
        tok = newToken(token.PLUS, l.curChar)
    case '{':
        tok = newToken(token.LBRA, l.curChar)
    case '}':
        tok = newToken(token.RBRA, l.curChar)
    case '-':
        tok = newToken(token.MINS, l.curChar)
    case '>':
        tok = newToken(token.GT, l.curChar)
    case '<':
        tok = newToken(token.LT, l.curChar)
    case '*':
        tok = newToken(token.ASTK, l.curChar)
    case '!':
        if l.peek() == '=' {
            tok = token.Token{Type: token.NEQ, 
                  Literal: string(l.curChar) + string(l.input[l.readPtr])}
            l.readChar();
        } else {
            tok = newToken(token.BANG, l.curChar)
        }
    case '/':
        tok = newToken(token.DIV, l.curChar)
    case 0:
        tok.Literal = ""
        tok.Type = token.EOF
    default:
        if isLetter(l.curChar){
            tok.Literal = l.readWithPredicate(isLetter)
            tok.Type = token.IdentifierLookup(tok.Literal)
            return tok
        } else if isDigit(l.curChar){
            tok.Literal = l.readWithPredicate(isDigit)
            tok.Type = token.INT
            return tok
        }else{
            tok = newToken(token.ERR, l.curChar)
        }
    }
    l.readChar()
    return tok
}

type predicate func(byte) bool

func isLetter(input byte) bool {
    return input >= 'a' && input <= 'z' || 
           input >= 'A' && input <= 'Z' || 
           input == '_'
}

func isDigit(input byte) bool {
    return (input >= '0' && input <= '9');
}

func (l* Lexer) readWithPredicate(fn predicate) string {
    startIdx := l.position
    for fn(l.curChar){
        l.readChar()
    }
    return l.input[startIdx:l.position]
}

func newToken(tokenType token.TokenType, char byte) token.Token {
    return token.Token{Type: tokenType,Literal:  string(char)}
}


func (l *Lexer) skipWhitespace(){
    for l.curChar == ' ' || 
       l.curChar == '\n' || 
       l.curChar == '\r' ||
       l.curChar == '\t' { 
        l.readChar()
    }
}

func (l *Lexer) peek() byte {
    if l.readPtr < len(l.input){
        return l.input[l.readPtr]
    }
    return 0
}
