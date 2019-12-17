// %はgoyaccの定義部
%{
package formula

import (
	"errors"
	"io"
	"text/scanner"
)

type Expression interface{}

type Token struct {
	Token   int
	Atomic string
}

type Atomic struct {
	Atomic string
}

type NotOpExpr struct {
	Operator rune
	Right    Expression
}

type BinOpExpr struct {
	Left     Expression
	Operator rune
	Right    Expression
}

%}

%union{
  token Token
  expr Expression
}

%type<expr> formula
%type<expr> expr and_expr or_expr not_expr imply_expr parenth_expr
%token<token> ATOMIC

%right '>'
%left '&' '|'
%right '~'

%%
// 以降はgoyaccの規則部.

formula
  : expr
  {
    $$ = $1
    yylex.(*Lexer).result = $$
  }

expr
	: ATOMIC
	{
		$$ = Atomic{Atomic: $1.Atomic}
	}
	| and_expr
	| or_expr
	| imply_expr
	| not_expr
	| parenth_expr

and_expr
	: expr '&' expr
	{
		$$ = BinOpExpr{Left: $1, Operator: '&', Right: $3}
	}

or_expr
	: expr '|' expr
	{
		$$ = BinOpExpr{Left: $1, Operator: '|', Right: $3}
	}

imply_expr
	: expr '>' expr
	{
		$$ = BinOpExpr{Left: $1, Operator: '>', Right: $3}
	}

not_expr
	: '~' expr
	{
		$$ = NotOpExpr{Operator: '~', Right: $2}
	}

parenth_expr
	: '(' expr ')'
	{
		$$ = $2
	}

%%
// 以降はgoyaccのユーザー定義部. Goで字句解析器(Lexer)と構文解析器(Parser)を記述する.

type Lexer struct {
	scanner.Scanner
	result Expression
	err	error
}

func (l *Lexer) Lex(lval /* lexer value */ *yySymType) int {
	token := int(l.Scan())
	if token == scanner.Ident || token == scanner.Int{
		token = ATOMIC
	}
	lval.token = Token{Token: token, Atomic: l.TokenText()}
	return token
}

func (l *Lexer) Error(e /* error */ string) {
	l.err = errors.New(e)
}

func Parse(r /* reader */ io.Reader) (Expression, error) {
	l := new(Lexer)
	l.Init(r)
	yyParse(l)
	if l.err != nil {
		return nil, l.err
	}
	return l.result, nil
}