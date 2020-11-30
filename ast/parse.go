package ast

import (
	"fmt"
	"io"

	"mibk.io/php/token"
)

// SyntaxError records an error and the position it occured on.
type SyntaxError struct {
	Line, Column int
	Err          error
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("line:%d:%d: %v", e.Line, e.Column, e.Err)
}

type parser struct {
	scan *token.Scanner

	err  error
	tok  token.Token
	prev token.Token
	alt  *token.Token // on backup
}

// Parse parses a single PHP file.
func Parse(r io.Reader) (*File, error) {
	p := &parser{scan: token.NewScanner(r)}
	p.next0() // init
	doc := p.parseFile()
	if p.err != nil {
		return nil, p.err
	}
	return doc, nil
}

func (p *parser) backup() {
	if p.alt != nil {
		panic("cannot backup twice")
	}
	p.alt = new(token.Token)
	*p.alt = p.tok
	p.tok = p.prev
}

func (p *parser) next0() {
	if p.alt != nil {
		p.tok, p.alt = *p.alt, nil
		return
	}
	p.tok = p.scan.Next()
}

// next is like next0 but skips whitespace.
func (p *parser) next() {
	p.prev = p.tok
	p.next0()
	p.consume(token.Whitespace)
}

func (p *parser) expect(typ token.Type) {
	if p.tok.Type != typ {
		p.errorf("expecting %v, found %v", typ, p.tok)
	}
	p.next()
}

func (p *parser) got(typ token.Type) bool {
	if p.tok.Type == typ {
		p.next()
		return true
	}
	return false
}

func (p *parser) consume(types ...token.Type) {
	if len(types) == 0 {
		panic("not token types to consume provided")
	}

	for ; len(types) > 0; types = types[1:] {
		if p.tok.Type == types[0] {
			p.next0()
		}
	}
}

func (p *parser) errorf(format string, args ...interface{}) {
	if p.err == nil {
		se := &SyntaxError{Err: fmt.Errorf(format, args...)}
		se.Line, se.Column = p.tok.Pos.Line, p.tok.Pos.Column
		p.err = se
	}
}

// The syntax comments roughly follow the notation as defined at
// https://golang.org/ref/spec#Notation.

// File = "<?php"
//        [ "namespace" Name ";" ]
//        { UseStmt }
//        { Decl } .
func (p *parser) parseFile() *File {
	file := new(File)
	p.expect(token.OpenTag)
	if p.got(token.Namespace) {
		file.Namespace = p.parseName()
		p.expect(token.Semicolon)
	}
	for p.tok.Type == token.Use {
		file.UseStmts = append(file.UseStmts, p.parseUseStmt())
	}
	// TODO: Avoid p.err == nil.
	for !p.got(token.EOF) && p.err == nil {
		file.Decls = append(file.Decls, p.parseDecl())
	}
	return file
}

// UseStmt = "use" Name ";" .
func (p *parser) parseUseStmt() *UseStmt {
	stmt := new(UseStmt)
	p.expect(token.Use)
	stmt.Name = p.parseName()
	p.expect(token.Semicolon)
	return stmt
}

// Decl = ConstDecl | ClassDecl .
func (p *parser) parseDecl() Decl {
	switch p.tok.Type {
	case token.Const:
		return p.parseConstDecl()
	case token.Class:
		return p.parseClassDecl()
	default:
		p.errorf("unexpected %v", p.tok)
		return nil
	}
}

// ConstDecl = "const" ident "=" Expr ";" .
func (p *parser) parseConstDecl() *ConstDecl {
	cons := new(ConstDecl)
	p.expect(token.Const)
	cons.Name = p.tok.Text
	p.expect(token.Ident)
	p.expect(token.Assign)
	cons.X = p.parseExpr()
	p.expect(token.Semicolon)
	return cons
}

// ClassDecl   = "class" "{" { ClassMember } "}" .
// ClassMember = Method | ConstDecl .
// Method      = "function" ident "(" ParamList [ "," ] ")" BlockStmt .
func (p *parser) parseClassDecl() *ClassDecl {
	class := new(ClassDecl)
	p.expect(token.Class)
	class.Name = p.tok.Text
	p.expect(token.Ident)
	p.expect(token.Lbrace)
	for !p.got(token.Rbrace) && p.err == nil {
		m := p.parseClassMember()
		class.Members = append(class.Members, m)
	}
	return class
}

func (p *parser) parseClassMember() Decl {
	switch p.tok.Type {
	default:
		p.errorf("unexpected %v, expecting class member", p.tok)
		return nil
	case token.Const:
		return p.parseConstDecl()
	case token.Function:
		p.next()
		m := new(Method)
		m.Name = p.tok.Text
		p.expect(token.Ident)
		p.expect(token.Lparen)
		m.Params = p.parseParamList()
		p.consume(token.Comma)
		p.expect(token.Rparen)
		m.Body = p.parseBlockStmt()
		return m
	}
}

// ParamList = Param { "," Param } .
// Param     = var .
func (p *parser) parseParamList() []*Param {
	var params []*Param
	for i := 0; p.tok.Type != token.Rparen; i++ {
		if i > 0 && !p.got(token.Comma) {
			break
		}
		if p.tok.Type != token.Var {
			break
		}
		params = append(params, &Param{Name: p.tok.Text})
		p.next()
	}
	return params
}

// BlockStmt = "{" { Stmt } "}" .
func (p *parser) parseBlockStmt() *BlockStmt {
	block := new(BlockStmt)
	p.expect(token.Lbrace)
	for {
		if p.got(token.Rbrace) || p.err != nil {
			return block
		}
		block.List = append(block.List, p.parseStmt())
	}
}

// Stmt = BlockStmt | UnknownStmt .
func (p *parser) parseStmt() Stmt {
	if p.tok.Type == token.Lbrace {
		return p.parseBlockStmt()
	}
	return p.parseUnknownStmt()
}

// Name = [ "\\" ] ident { "\\" ident } .
func (p *parser) parseName() *Name {
	id := new(Name)
	if p.got(token.Backslash) {
		id.Global = true
	}
	for {
		id.Parts = append(id.Parts, p.tok.Text)
		p.expect(token.Ident)
		if !p.got(token.Backslash) {
			break
		}
	}
	return id
}

// UnknownStmt = /* pretty much anything */ ( ";" | BlockStmt ) .
func (p *parser) parseUnknownStmt() *UnknownStmt {
	stmt := new(UnknownStmt)
	// TODO: EOF or ?>
	for {
		switch p.tok.Type {
		case token.EOF, token.Rbrace:
			p.errorf("unexpected %v", p.tok)
			return nil
		case token.Semicolon:
			if len(stmt.Toks) == 0 {
				p.errorf("unexpected %v", p.tok)
			}
			p.next()
			return stmt
		case token.Lbrace:
			stmt.Body = p.parseBlockStmt()
			return stmt
		default:
			stmt.Toks = append(stmt.Toks, p.tok)
			p.next0()
		}
	}
}

// Expr = UnknownExpr .
func (p *parser) parseExpr() Expr { return p.parseUnknownExpr() }

// UnknownExpr = /* anything except ";" */ .
func (p *parser) parseUnknownExpr() *UnknownExpr {
	x := new(UnknownExpr)
	// TODO: EOF or ?>
	for p.tok.Type != token.Semicolon && p.tok.Type != token.EOF {
		x.Toks = append(x.Toks, p.tok)
		p.next0()
	}
	if len(x.Toks) == 0 {
		p.errorf("unexpected %v", p.tok)
	}
	return x
}
