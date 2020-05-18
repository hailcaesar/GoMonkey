package parser

import (
	"fmt"
	"gomonkey/ast"
	"gomonkey/lexer"
	"gomonkey/token"
	"strconv"
)

type Parser struct {
	lexer *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// Pratt Parser Implementation
type prefixParseFn func() ast.Expression
type infixParseFn func(ast.Expression) ast.Expression

func New(l *lexer.Lexer) *Parser {
	parser := &Parser{lexer: l, errors: []string{}}
	parser.advance()
	parser.advance()

	parser.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	parser.prefixParseFns[token.BANG] = parser.parsePrefixExpression
	parser.prefixParseFns[token.FALS] = parser.parseBoolean
	parser.prefixParseFns[token.FNCT] = parser.parseFunction
	parser.prefixParseFns[token.IDN] = parser.parseIdentifier
	parser.prefixParseFns[token.IF] = parser.parseIfExpression
	parser.prefixParseFns[token.INT] = parser.parseIntegerLiteral
	parser.prefixParseFns[token.LPAR] = parser.parseGroupedExpression
	parser.prefixParseFns[token.MINS] = parser.parsePrefixExpression
	parser.prefixParseFns[token.TRUE] = parser.parseBoolean

	parser.infixParseFns = make(map[token.TokenType]infixParseFn)
	parser.infixParseFns[token.ASTK] = parser.parseInfixExpression
	parser.infixParseFns[token.DIV] = parser.parseInfixExpression
	parser.infixParseFns[token.EQ] = parser.parseInfixExpression
	parser.infixParseFns[token.GT] = parser.parseInfixExpression
	parser.infixParseFns[token.LPAR] = parser.parseCallExpression
	parser.infixParseFns[token.LT] = parser.parseInfixExpression
	parser.infixParseFns[token.MINS] = parser.parseInfixExpression
	parser.infixParseFns[token.NEQ] = parser.parseInfixExpression
	parser.infixParseFns[token.PLUS] = parser.parseInfixExpression
	return parser
}

func (p *Parser) advance() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) ParseCode() *ast.Code {
	code := &ast.Code{}
	code.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		statement := p.parseStatement()
		if statement != nil {
			code.Statements = append(code.Statements, statement)
		}
		p.advance()
	}
	return code
}

// 5 + 7 * 8
// left = 5
// infix = +

//precedence = +
// left = 7
// infix = *

//precedence = *
//left = 8
//semicolon --> return

//infix.left = 7, infix.op = *, infix.right = 8
//infix. left = 5, infix.op = +, infix.right = (above)

// 5 * 7 + 8
// left = 5
// infix = *

//precedence = *
// left = 7
// + has lower precedence than * so return

//left = infix.left = 5, infix.op = *, infix.right = 7
// + has higher priroity than 'LOWEST'
// left = infix(left)
// infix.operator = +
// infix.right = parseExpression()
// 						left = 8
//						semicolon so return
// left =

// A + B
// A + (B * (C % D))

func (p *Parser) parseExpression(precedence int) ast.Expression {
	//	defer untrace(trace("parseExpression"))
	prefixFn := p.prefixParseFns[p.curToken.Type]
	if prefixFn == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExpression := prefixFn()
	for p.peekToken.Type != token.SCLN && precedence < p.peekPrecedence() {
		infixFn := p.infixParseFns[p.peekToken.Type]
		if infixFn == nil {
			return leftExpression
		}

		p.advance()
		newExpression := infixFn(leftExpression)
		leftExpression = newExpression //just to make it clear the returned value is whole new expression
	}
	return leftExpression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	//defer untrace(trace("parseInfixExpression"))
	infixExp := &ast.InfixExpression{}
	infixExp.Token = p.curToken
	infixExp.Operator = p.curToken.Literal
	infixExp.Left = left

	precedence := precedences[p.curToken.Type]
	p.advance()
	infixExp.Right = p.parseExpression(precedence)

	return infixExp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	//defer untrace(trace("parsePrefixExpression"))
	prefixExp := &ast.PrefixExpression{}
	prefixExp.Token = p.curToken
	prefixExp.Operator = p.curToken.Literal

	p.advance()
	prefixExp.Right = p.parseExpression(PREFIX)

	return prefixExp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.advance()

	exp := p.parseExpression(LOWEST)

	if !p.advanceIf(token.RPAR) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAR) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAR) {
		return nil
	}

	if !p.expectPeek(token.LBRA) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRA) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseIfExpression2() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if !p.advanceIf(token.LPAR) {
		return nil
	}
	p.advance()

	expression.Condition = p.parseExpression(LOWEST)
	if !p.advanceIf(token.RPAR) {
		return nil
	}
	if !p.advanceIf(token.LBRA) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekToken.Type == token.ELSE {
		p.advance()
		if p.peekToken.Type != token.LBRA {
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

func (p *Parser) parseFunction() ast.Expression {
	fnLit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAR) {
		return nil
	}

	fnLit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRA) {
		return nil
	}

	fnLit.Body = p.parseBlockStatement()

	return fnLit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}
	if p.peekTokenIs(token.RPAR) {
		p.nextToken()
		return nil
	}

	p.nextToken()

	identifier := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, identifier)

	for p.peekTokenIs(token.COM) {
		p.nextToken()
		p.nextToken()
		identifier := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, identifier)
	}

	if !p.expectPeek(token.RPAR) {
		return nil
	}

	return identifiers

}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	intlit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("Error: Integer not valid [Value=%q]", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	intlit.Value = value
	return intlit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curToken.Type == token.TRUE}
}

func (p *Parser) advanceIf(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.advance()
		return true
	}
	p.addPeekError(t)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addPeekError(tok token.TokenType) {
	err := fmt.Sprintf("Error: Exepected '%s' token [actual = '%s']", tok, p.peekToken.Type)
	p.errors = append(p.errors, err)
}

var precedences = map[token.TokenType]int{
	token.EQ:   EQUALS,
	token.NEQ:  EQUALS,
	token.LT:   LESSGREATER,
	token.GT:   LESSGREATER,
	token.PLUS: SUM,
	token.MINS: SUM,
	token.DIV:  PRODUCT,
	token.ASTK: PRODUCT,
	token.LPAR: CALL,
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peekToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.curToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

/*  ---------------------------------------------------------------------------- */
/*  ---------------------------- Parse Statements ------------------------------ */
/*  ---------------------------------------------------------------------------- */

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RET:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() ast.Statement {
	let := &ast.LetStatement{}
	let.Token = p.curToken

	ok := p.advanceIf(token.IDN)
	if !ok {
		return nil
	}

	let.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	ok = p.advanceIf(token.AGMT)
	if !ok {
		return nil
	}

	p.nextToken()
	let.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SCLN) {
		p.nextToken()
	}

	return let
}

func (p *Parser) parseReturnStatement() ast.Statement {
	ret := &ast.ReturnStatement{}
	ret.Token = p.curToken
	p.advance()

	ret.Value = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SCLN) {
		p.nextToken()
	}

	return ret
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	exp := &ast.ExpressionStatement{}
	exp.Token = p.curToken

	exp.Expression = p.parseExpression(LOWEST)

	if p.peekToken.Type == token.SCLN {
		p.advance()
	}

	return exp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.advance()
	for p.curToken.Type != token.RBRA && p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.advance()
	}
	return block
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}
	if p.peekTokenIs(token.RPAR) {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COM) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(token.RPAR) {
		return nil
	}

	return args
}
