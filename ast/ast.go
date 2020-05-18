package ast

import (
	"bytes"
	"gomonkey/token"
	"strings"
)

/*  ----------------------------------------------------------- */
/*  --- Interfaces -------------------------------------------- */
/*  ----------------------------------------------------------- */

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Code struct {
	Statements []Statement
}

func (c *Code) TokenLiteral() string {
	if len(c.Statements) > 0 {
		return c.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (c *Code) String() string {
	var output bytes.Buffer
	for _, s := range c.Statements {
		output.WriteString(s.String())
	}
	return output.String()
}

/*  ----------------------------------------------------------- */
/*  --- Statements -------------------------------------------- */
/*  ----------------------------------------------------------- */

/* Anything other than 'let' or 'return' is part of an ExpressionStatement */
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (e *ExpressionStatement) statementNode()       {}
func (e *ExpressionStatement) TokenLiteral() string { return e.Token.Literal }
func (e *ExpressionStatement) String() string {
	if e.Expression != nil {
		return e.Expression.String()
	}
	return ""
}

/*
  Variable declarations
*/
type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var output bytes.Buffer
	output.WriteString(ls.TokenLiteral() + " ")
	output.WriteString(ls.Name.String())
	output.WriteString(" = ")
	if ls.Value != nil {
		output.WriteString(ls.Value.String())
	}
	output.WriteString(";")
	return output.String()
}

/*
  Return keyword + optional expression
  Eg.  return a + b,  return 5 * 10, return fn(a,b){ a + b }
*/
type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (r *ReturnStatement) statementNode()       {}
func (r *ReturnStatement) TokenLiteral() string { return r.Token.Literal }
func (r *ReturnStatement) String() string {
	var output bytes.Buffer
	output.WriteString(r.TokenLiteral() + " ")

	if r.Value != nil {
		output.WriteString(r.Value.String())
	}

	output.WriteString(";")
	return output.String()
}

/*
  Collection of statements that occur in an if block
*/
type BlockStatement struct {
	Token      token.Token // '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

/*  ----------------------------------------------------------- */
/*  --- Expressions ------------------------------------------- */
/*  ----------------------------------------------------------- */

/* Variable names */
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

/* Expressions that have a left and right operands */
type InfixExpression struct {
	Token    token.Token //  Operator token (e.g. +)
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var output bytes.Buffer
	output.WriteString("(")
	output.WriteString(oe.Left.String())
	output.WriteString(" " + oe.Operator + " ")
	output.WriteString(oe.Right.String())
	output.WriteString(")")
	return output.String()
}

/* Integers only, doesn't support floats/doubles */
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

/*
   Expressions meant for operators that do not have a left expression
   Eg.  -5, !ok, -count  */
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var output bytes.Buffer
	output.WriteString("(")
	output.WriteString(pe.Operator)
	output.WriteString(pe.Right.String())
	output.WriteString(")")
	return output.String()
}

/* True and False */
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

/*
 If statements
*/
type IfExpression struct {
	Token       token.Token // 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

/*
  Function declarations
*/
type FunctionLiteral struct {
	Token      token.Token // fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

/*
  Function calls
  foo(a, b, c),   bar(a, 5*8+9, z)
*/
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}
