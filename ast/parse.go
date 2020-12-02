package ast

import (
	"fmt"
	"io"
	"strings"

	"mibk.io/php/token"
	"mibk.io/phpdoc"
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
	p.consume(token.Whitespace, token.Comment, token.Whitespace)
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
		panic("no token types to consume provided")
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

// Decl = ConstDecl | VarDecl | FuncDecl | ClassDecl .
func (p *parser) parseDecl() Decl {
	doc := p.parsePHPDoc()
	switch p.tok.Type {
	case token.Const:
		return p.parseConstDecl(doc)
	case token.Var:
		return p.parseVarDecl(doc)
	case token.Function:
		return p.parseFuncDecl(doc)
	case token.Class:
		return p.parseClassDecl(doc)
	default:
		p.errorf("unexpected %v", p.tok)
		return nil
	}
}

// ConstDecl = "const" ident "=" Expr ";" .
func (p *parser) parseConstDecl(doc *phpdoc.Block) *ConstDecl {
	cons := new(ConstDecl)
	cons.Doc = doc
	p.expect(token.Const)
	cons.Name = p.tok.Text
	p.expect(token.Ident)
	p.expect(token.Assign)
	cons.X = p.parseExpr()
	p.expect(token.Semicolon)
	return cons
}

// VarDecl = var [ "=" Expr ] ";" .
func (p *parser) parseVarDecl(doc *phpdoc.Block) *VarDecl {
	cons := new(VarDecl)
	cons.Doc = doc
	cons.Name = p.tok.Text
	p.expect(token.Var)
	if p.got(token.Assign) {
		cons.X = p.parseExpr()
	}
	p.expect(token.Semicolon)
	return cons
}

// FuncDecl = "function" ident ParamList BlockStmt .
func (p *parser) parseFuncDecl(doc *phpdoc.Block) *FuncDecl {
	fn := new(FuncDecl)
	fn.Doc = doc
	p.expect(token.Function)
	fn.Name = p.tok.Text
	p.expect(token.Ident)
	fn.Params = p.parseParamList()
	if p.got(token.Colon) {
		fn.Result = p.parseName()
	}
	fn.Body = p.parseBlockStmt()
	return fn
}

// ParamList = "(" [ Param { "," Param } [ "," ] ] ")" .
// Param     = [ Name ] var .
func (p *parser) parseParamList() []*Param {
	var params []*Param
	p.expect(token.Lparen)
	for !p.got(token.Rparen) {
		par := new(Param)
		if p.tok.Type == token.Ident || p.tok.Type == token.Backslash {
			// TODO: Use better approach.
			par.Type = p.parseName()
		}
		par.Name = p.tok.Text
		p.expect(token.Var)
		params = append(params, par)
		if p.got(token.Rparen) {
			break
		}
		p.expect(token.Comma)
	}
	return params
}

// ClassDecl   = "class" [ "extends" Name ] "{" { UseStmt } { ClassMember } "}" .
// ClassMember = ConstDecl | FuncDecl .
func (p *parser) parseClassDecl(doc *phpdoc.Block) *ClassDecl {
	class := new(ClassDecl)
	class.Doc = doc
	p.expect(token.Class)
	class.Name = p.tok.Text
	p.expect(token.Ident)
	if p.got(token.Extends) {
		class.Extends = p.parseName()
	}
	p.expect(token.Lbrace)
	for p.tok.Type == token.Use {
		class.Traits = append(class.Traits, p.parseUseStmt())
	}
	for !p.got(token.Rbrace) && p.err == nil {
		m := p.parseClassMember()
		class.Members = append(class.Members, m)
	}
	return class
}

func (p *parser) parseClassMember() *ClassMember {
	m := new(ClassMember)
	m.Doc = p.parsePHPDoc()
	m.Vis = p.parseVisibility()
	switch p.tok.Type {
	default:
		p.errorf("unexpected %v, expecting %v or %v", p.tok, token.Const, token.Function)
		return nil
	case token.Const:
		m.Decl = p.parseConstDecl(nil)
	case token.Var:
		m.Decl = p.parseVarDecl(nil)
	case token.Function:
		m.Decl = p.parseFuncDecl(nil)
	}
	return m
}

func (p *parser) parseVisibility() Vis {
	var v Vis
	switch p.tok.Type {
	default:
		return DefaultVis
	case token.Public:
		v = Public
	case token.Protected:
		v = Protected
	case token.Private:
		v = Private
	}
	p.next()
	return v
}

func (p *parser) parsePHPDoc() *phpdoc.Block {
	if p.tok.Type != token.DocComment {
		return nil
	}
	doc, err := phpdoc.Parse(strings.NewReader(p.tok.Text))
	if err != nil {
		// TODO: do not panic
		panic(err)
	}
	p.next()
	return doc
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
