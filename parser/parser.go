package parser

import (
	"fmt"
	"gomonkey/ast"
	"gomonkey/lexer"
	tk "gomonkey/token"
	"strconv"
)

type Parser struct {
	lexer *lexer.Lexer

	cur  tk.Token
	next tk.Token

	errors []string

	prefixParseFns map[tk.TokenType]prefixParseFn
	infixParseFns  map[tk.TokenType]infixParseFn
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
	parser.registerCallbacks()

	return parser
}

func (p *Parser) registerCallbacks() {
	p.prefixParseFns = make(map[tk.TokenType]prefixParseFn)
	p.prefixParseFns[tk.BANG] = p.parsePrefixExpression
	p.prefixParseFns[tk.FALS] = p.parseBoolean
	p.prefixParseFns[tk.FNCT] = p.parseFunction
	p.prefixParseFns[tk.IDN] = p.parseIdentifier
	p.prefixParseFns[tk.IF] = p.parseIfExpression
	p.prefixParseFns[tk.INT] = p.parseIntegerLiteral
	p.prefixParseFns[tk.LPAR] = p.parseGroupedExpression
	p.prefixParseFns[tk.MINS] = p.parsePrefixExpression
	p.prefixParseFns[tk.TRUE] = p.parseBoolean

	p.infixParseFns = make(map[tk.TokenType]infixParseFn)
	p.infixParseFns[tk.ASTK] = p.parseInfixExpression
	p.infixParseFns[tk.DIV] = p.parseInfixExpression
	p.infixParseFns[tk.EQ] = p.parseInfixExpression
	p.infixParseFns[tk.GT] = p.parseInfixExpression
	p.infixParseFns[tk.LPAR] = p.parseCallExpression
	p.infixParseFns[tk.LT] = p.parseInfixExpression
	p.infixParseFns[tk.MINS] = p.parseInfixExpression
	p.infixParseFns[tk.NEQ] = p.parseInfixExpression
	p.infixParseFns[tk.PLUS] = p.parseInfixExpression
}

var precedences = map[tk.TokenType]int{
	tk.EQ:   EQUALS,
	tk.NEQ:  EQUALS,
	tk.LT:   LESSGREATER,
	tk.GT:   LESSGREATER,
	tk.PLUS: SUM,
	tk.MINS: SUM,
	tk.DIV:  PRODUCT,
	tk.ASTK: PRODUCT,
	tk.LPAR: CALL,
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.next.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.cur.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) ParseCode() *ast.Code {
	code := &ast.Code{}
	code.Statements = []ast.Statement{}
	for p.cur.Type != tk.EOF {
		statement := p.parseStatement()
		if statement != nil {
			code.Statements = append(code.Statements, statement)
		}
		p.advance()
	}
	return code
}

/*  ----------------------------------------------------------- */
/*  --- Parse Expressions ------------------------------------- */
/*  ----------------------------------------------------------- */

// Function Calls
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.cur, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}
	if p.nextIs(tk.RPAR) {
		p.advance()
		return args
	}
	p.advance()
	args = append(args, p.parseExpression(LOWEST))
	for p.nextIs(tk.COM) {
		p.advance()
		p.advance()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.advanceIfNextIs(tk.RPAR) {
		return nil
	}

	return args
}

/*
 *
 * Main Recursive parsing function
 *
 */
func (p *Parser) parseExpression(precedence int) ast.Expression {
	//	defer untrace(trace("parseExpression"))
	prefixFn := p.prefixParseFns[p.cur.Type]
	if prefixFn == nil {
		p.noPrefixParseFnError(p.cur.Type)
		return nil
	}

	leftExpression := prefixFn()
	for p.next.Type != tk.SCLN && precedence < p.peekPrecedence() {
		infixFn := p.infixParseFns[p.next.Type]
		if infixFn == nil {
			return leftExpression
		}

		p.advance()
		newExpression := infixFn(leftExpression)
		leftExpression = newExpression //just to make it clear the returned value is whole new expression
	}
	return leftExpression
}

// Infix operators
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	//defer untrace(trace("parseInfixExpression"))
	infixExp := &ast.InfixExpression{}
	infixExp.Token = p.cur
	infixExp.Operator = p.cur.Literal
	infixExp.Left = left

	precedence := precedences[p.cur.Type]
	p.advance()
	infixExp.Right = p.parseExpression(precedence)

	return infixExp
}

// Prefix operators
func (p *Parser) parsePrefixExpression() ast.Expression {
	//defer untrace(trace("parsePrefixExpression"))
	prefixExp := &ast.PrefixExpression{}
	prefixExp.Token = p.cur
	prefixExp.Operator = p.cur.Literal

	p.advance()
	prefixExp.Right = p.parseExpression(PREFIX)

	return prefixExp
}

// Expressions surround by parenthesis.... ie  c * (a + b)
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.advance()

	exp := p.parseExpression(LOWEST)

	if !p.advanceIf(tk.RPAR) {
		return nil
	}

	return exp
}

// If statmenets
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.cur}

	if !p.advanceIfNextIs(tk.LPAR) {
		return nil
	}

	p.advance()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.advanceIfNextIs(tk.RPAR) {
		return nil
	}

	if !p.advanceIfNextIs(tk.LBRA) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.nextIs(tk.ELSE) {
		p.advance()

		if !p.advanceIfNextIs(tk.LBRA) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseIfExpression2() ast.Expression {
	expression := &ast.IfExpression{Token: p.cur}
	if !p.advanceIf(tk.LPAR) {
		return nil
	}
	p.advance()

	expression.Condition = p.parseExpression(LOWEST)
	if !p.advanceIf(tk.RPAR) {
		return nil
	}
	if !p.advanceIf(tk.LBRA) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.next.Type == tk.ELSE {
		p.advance()
		if p.next.Type != tk.LBRA {
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

/* Function declarations */
func (p *Parser) parseFunction() ast.Expression {
	fnLit := &ast.FunctionLiteral{Token: p.cur}

	if !p.advanceIfNextIs(tk.LPAR) {
		return nil
	}

	fnLit.Parameters = p.parseFunctionParameters()

	if !p.advanceIfNextIs(tk.LBRA) {
		return nil
	}

	fnLit.Body = p.parseBlockStatement()

	return fnLit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}
	if p.nextIs(tk.RPAR) {
		p.advance()
		return nil
	}

	p.advance()

	identifier := &ast.Identifier{Token: p.cur, Value: p.cur.Literal}
	identifiers = append(identifiers, identifier)

	for p.nextIs(tk.COM) {
		p.advance()
		p.advance()
		identifier := &ast.Identifier{Token: p.cur, Value: p.cur.Literal}
		identifiers = append(identifiers, identifier)
	}

	if !p.advanceIfNextIs(tk.RPAR) {
		return nil
	}

	return identifiers

}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.cur, Value: p.cur.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	intlit := &ast.IntegerLiteral{Token: p.cur}

	value, err := strconv.ParseInt(p.cur.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("Error: Integer not valid [Value=%q]", p.cur.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	intlit.Value = value
	return intlit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.cur, Value: p.cur.Type == tk.TRUE}
}

func (p *Parser) advanceIf(t tk.TokenType) bool {
	if p.next.Type == t {
		p.advance()
		return true
	}
	p.addPeekError(t)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addPeekError(tok tk.TokenType) {
	err := fmt.Sprintf("Error: Exepected '%s' token [actual = '%s']", tok, p.next.Type)
	p.errors = append(p.errors, err)
}

func (p *Parser) noPrefixParseFnError(t tk.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) advance() {
	p.cur = p.next
	p.next = p.lexer.NextToken()
}

func (p *Parser) curIs(t tk.TokenType) bool {
	return p.cur.Type == t
}

func (p *Parser) nextIs(t tk.TokenType) bool {
	return p.next.Type == t
}

func (p *Parser) advanceIfNextIs(t tk.TokenType) bool {
	if p.nextIs(t) {
		p.advance()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t tk.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.next.Type)
	p.errors = append(p.errors, msg)
}

/*  ----------------------------------------------------------- */
/*  --- Parse Statement --------------------------------------- */
/*  ----------------------------------------------------------- */

func (p *Parser) parseStatement() ast.Statement {
	switch p.cur.Type {
	case tk.LET:
		return p.parseLetStatement()
	case tk.RET:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() ast.Statement {
	let := &ast.LetStatement{}
	let.Token = p.cur

	ok := p.advanceIf(tk.IDN)
	if !ok {
		return nil
	}

	let.Name = &ast.Identifier{Token: p.cur, Value: p.cur.Literal}

	ok = p.advanceIf(tk.AGMT)
	if !ok {
		return nil
	}

	p.advance()
	let.Value = p.parseExpression(LOWEST)

	if p.nextIs(tk.SCLN) {
		p.advance()
	}

	return let
}

func (p *Parser) parseReturnStatement() ast.Statement {
	ret := &ast.ReturnStatement{}
	ret.Token = p.cur
	p.advance()

	ret.Value = p.parseExpression(LOWEST)

	for p.nextIs(tk.SCLN) {
		p.advance()
	}

	return ret
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	exp := &ast.ExpressionStatement{}
	exp.Token = p.cur

	exp.Expression = p.parseExpression(LOWEST)

	if p.next.Type == tk.SCLN {
		p.advance()
	}

	return exp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.cur}
	block.Statements = []ast.Statement{}
	p.advance()
	for p.cur.Type != tk.RBRA && p.cur.Type != tk.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.advance()
	}
	return block
}
