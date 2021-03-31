/*
 * @Author: dongzhzheng
 * @Date: 2021-03-31 15:58:49
 * @LastEditTime: 2021-03-31 16:01:40
 * @LastEditors: dongzhzheng
 * @FilePath: /go-interpreter-demo/main.go
 * @Description:
 */
package main

import (
	"fmt"
	"log"
	"strings"
	"text/scanner"
)

/*
syntax:

statement:  ident "=" expression
expression: term   { ("+" | "-") expression }
term:       factor { ("*" | "/") term }
factor:     NUMBER | "(" expression ")" | - factor | ident
ident:      [A-Za-z_]*
*/

// Token ...
type Token struct {
	Pos  int
	Val  string
	Type rune
}

// Interpreter ...
type Interpreter struct {
	Sympol map[rune]float64

	tokens     []Token
	tokenIndex int
}

// Init ...
func (interp *Interpreter) Init() {
	interp.Sympol = map[rune]float64{}
	interp.tokens = []Token{{
		Pos:  0,
		Val:  "",
		Type: 0,
	}}
}

// AddToken ...
func (interp *Interpreter) AddToken(t Token) {
	interp.tokens = append(interp.tokens, t)
}

// Column ...
func (interp *Interpreter) Column() int {
	return interp.tokens[interp.tokenIndex].Pos
}

// Pos ...
func (interp *Interpreter) Pos() int {
	return interp.tokenIndex
}

// TokenText ...
func (interp *Interpreter) TokenText() string {
	return interp.tokens[interp.tokenIndex].Val
}

// Scan ...
func (interp *Interpreter) Scan() rune {
	interp.tokenIndex++
	if interp.tokenIndex >= len(interp.tokens) {
		interp.tokenIndex = len(interp.tokens) - 1
	}
	return interp.tokens[interp.tokenIndex].Type
}

// Back ...
func (interp *Interpreter) Back() {
	interp.tokenIndex--
}

type Expr interface{}

// BinExprAST ...
type BinExprAST struct {
	Op       string
	LHS, RHS Expr
}

// String ...
func (b *BinExprAST) String() {
	fmt.Printf("op(%s) lhs(%v) rhs(%v)\n", b.Op, b.LHS, b.RHS)
}

// UnaryExprAST ...
type UnaryExprAST struct {
	Op  string
	RHS Expr
}

type NumExprAST struct {
	Val  string
	Type int
}

// String ...
func (n *NumExprAST) String() {
	fmt.Printf("val(%s) type(%d)\n", n.Val, n.Type)
}

var (
	interpreter Interpreter
)

func myparser(s *scanner.Scanner) {
	interpreter.Init()

	for i := s.Scan(); i != scanner.EOF; i = s.Scan() {
		interpreter.AddToken(Token{
			Pos:  s.Pos().Column,
			Val:  s.TokenText(),
			Type: i,
		})
	}
	interpreter.AddToken(Token{
		Pos:  s.Pos().Column,
		Val:  "",
		Type: scanner.EOF,
	})

	b, err := parseStmt(&interpreter)
	if err != nil {
		log.Fatalf("parseStmt err(%v)", err)
	}

	eval(b)
}

func parseStmt(s *Interpreter) (Expr, error) {
	tok := s.Scan()
	if tok == scanner.EOF {
		return nil, fmt.Errorf("unexcept EOF")
	}

	if tok == scanner.Ident {
		ident := s.TokenText()
		tok = s.Scan()
		if tok == scanner.EOF {
			return nil, fmt.Errorf("unexcept EOF")
		}

		if s.TokenText() == "=" {
			rhs, err := parseExpr(s)
			if err != nil {
				return nil, fmt.Errorf("parseExpr err(%w)", err)
			}

			tok = s.Scan()
			if tok != scanner.EOF {
				return nil, fmt.Errorf("unexcept next token(\"%s\")", s.TokenText())
			}

			return &BinExprAST{
				Op: "=",
				LHS: &NumExprAST{
					Val:  ident,
					Type: scanner.Ident,
				},
				RHS: rhs,
			}, nil
		}
		return nil, fmt.Errorf("pos(%d) invalid token(\"%s\") need \"=\"", s.Column(), s.TokenText())
	}

	return nil, fmt.Errorf("pos(%d) invalid token(\"%s\")", s.Column(), s.TokenText())
}

func parseExpr(s *Interpreter) (Expr, error) {
	lhs, err := parseTerm(s)
	if err != nil {
		return nil, fmt.Errorf("parseTerm err(%w)", err)
	}

	tok := s.Scan()
	if tok == scanner.EOF {
		return lhs, nil
	}

	op := ""
	switch s.TokenText() {
	case "+":
		fallthrough
	case "-":
		op = s.TokenText()
		rhs, err := parseExpr(s)
		if err != nil {
			return nil, fmt.Errorf("parseExpr err(%w)", err)
		}
		return &BinExprAST{
			Op:  op,
			LHS: lhs,
			RHS: rhs,
		}, nil
	}

	s.Back()
	return lhs, nil
}

func parseTerm(s *Interpreter) (Expr, error) {
	lhs, err := parseFactor(s)
	if err != nil {
		return nil, fmt.Errorf("parseFactor err(%w)", err)
	}

	tok := s.Scan()
	if tok == scanner.EOF {
		return lhs, nil
	}

	op := ""
	switch s.TokenText() {
	case "*":
		fallthrough
	case "/":
		op = s.TokenText()
		rhs, err := parseTerm(s)
		if err != nil {
			return nil, fmt.Errorf("parseTerm err(%w)", err)
		}
		return &BinExprAST{
			Op:  op,
			LHS: lhs,
			RHS: rhs,
		}, nil
	}

	s.Back()
	return lhs, nil
}

func parseFactor(s *Interpreter) (Expr, error) {
	tok := s.Scan()
	if tok == scanner.Int {
		return &NumExprAST{
			Val:  s.TokenText(),
			Type: scanner.Int,
		}, nil
	}

	if tok == scanner.Float {
		return &NumExprAST{
			Val:  s.TokenText(),
			Type: scanner.Float,
		}, nil
	}

	switch s.TokenText() {
	case "(":
		expr, err := parseExpr(s)
		if err != nil {
			return nil, fmt.Errorf("parseExpr err(%w)", err)
		}
		_ = s.Scan()
		if s.TokenText() != ")" {
			return nil, fmt.Errorf("pos(%d) invalid token(\"%s\") need \")\"", s.Column(), s.TokenText())
		}
		return expr, nil
	case "-":
		rhs, err := parseFactor(s)
		if err != nil {
			return nil, fmt.Errorf("parseFactor err(%w)", err)
		}
		return &UnaryExprAST{
			Op:  "-",
			RHS: rhs,
		}, nil
	}

	if tok == scanner.Ident {
		return &NumExprAST{
			Val:  s.TokenText(),
			Type: scanner.Ident,
		}, nil
	}

	return nil, fmt.Errorf("pos(%d) invalid token(\"%s\") need \"number | ( | - | ident\"",
		s.Column(), s.TokenText())
}

func eval(b Expr) {
	switch b.(type) {
	case BinExprAST:
	case UnaryExprAST:
	case NumExprAST:
	}
}

func evalBinExprAST(b *BinExprAST) {
	//
}

func evalUnaryExprAST(u *UnaryExprAST) {
	//
}

func evalNumExprAST(u *NumExprAST) {
	//
}

func main() {
	const src = `B=(8/10)+A+9`
	const filename = "formula"
	var sc scanner.Scanner
	sc.Init(strings.NewReader(src))
	sc.Filename = filename
	myparser(&sc)
}
