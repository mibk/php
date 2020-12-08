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

// Parse parses a single PHP file. If an error occurs while parsing
// (except io errors), the returned error will be of type *SyntaxError.
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
	if p.tok.Type == token.EOF {
		return
	}
	if p.alt != nil {
		p.tok, p.alt = *p.alt, nil
		return
	}
	p.tok = p.scan.Next()
	if p.tok.Type == token.EOF && p.err == nil {
		err := p.scan.Err()
		if se, ok := err.(*token.ScanError); ok {
			// Make sure we always return *SyntaxError.
			p.err = &SyntaxError{
				Line:   se.Pos.Line,
				Column: se.Pos.Column,
				Err:    se.Err,
			}
		} else if err != nil {
			p.errorf("scan: %v", err)
		}
	}
}

// next is like next0 but skips whitespace.
func (p *parser) next() {
	p.prev = p.tok
	p.next0()
	for p.tok.Type == token.Whitespace || p.tok.Type == token.Comment {
		p.next0()
	}
}

func (p *parser) expect(typ token.Type) string {
	if p.tok.Type != typ {
		p.errorf("expecting %v, found %v", typ, p.tok)
	}
	text := p.tok.Text
	p.next()
	return text
}

func (p *parser) got(typ token.Type) bool {
	if p.tok.Type == typ {
		p.next()
		return true
	}
	return false
}

func (p *parser) until(typ token.Type) bool {
	return p.tok.Type != typ && p.tok.Type != token.EOF
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
		p.tok.Type = token.EOF
		se := &SyntaxError{Err: fmt.Errorf(format, args...)}
		se.Line, se.Column = p.tok.Pos.Line, p.tok.Pos.Column
		p.err = se
	}
}

// The syntax comments roughly follow the notation as defined at
// https://golang.org/ref/spec#Notation.

// File = "<?php"
//        { Pragma }
//        [ "namespace" Name ";" ]
//        { UseStmt }
//        { TopLevelStmt } .
func (p *parser) parseFile() *File {
	file := new(File)
	p.expect(token.OpenTag)
	// TODO: Allow on other places in a file?
	file.Pragmas = p.parsePragmas()
	if p.got(token.Namespace) {
		file.Namespace = p.parseName()
		p.expect(token.Semicolon)
	}
	for p.tok.Type == token.Use {
		file.UseStmts = append(file.UseStmts, p.parseUseStmt())
	}
	for !p.got(token.EOF) {
		file.Stmts = append(file.Stmts, p.parseTopLevelStmt())
	}
	return file
}

// Pragma = "declare" "(" Name "=" BasicLit ")" ";" .
func (p *parser) parsePragmas() []*Pragma {
	var pragmas []*Pragma
	for p.got(token.Declare) {
		d := new(Pragma)
		p.expect(token.Lparen)
		d.Name = p.expect(token.Ident)
		p.expect(token.Assign)
		d.Value = p.parseBasicLit()
		p.expect(token.Rparen)
		// TODO: Also parse body?
		p.expect(token.Semicolon)
		pragmas = append(pragmas, d)
	}
	return pragmas
}

// UseStmt = "use" Name [ "as" ident ] ";" .
func (p *parser) parseUseStmt() *UseStmt {
	stmt := new(UseStmt)
	p.expect(token.Use)
	stmt.Name = p.parseName()
	if p.got(token.As) {
		stmt.Alias = p.expect(token.Ident)
	}
	p.expect(token.Semicolon)
	return stmt
}

// TopLevelStmt = ConstDecl |
//                VarDecl |
//                FuncDecl |
//                ClassDecl |
//                InterfaceDecl |
//                Stmt .
func (p *parser) parseTopLevelStmt() Stmt {
	doc := p.parsePHPDoc()
	switch p.tok.Type {
	case token.Const:
		return p.parseConstDecl(doc)
	case token.Var:
		return p.parseVarDecl(doc, false)
	case token.Function:
		return p.parseFuncDecl(doc, false)
	case token.Class, token.Abstract:
		return p.parseClassDecl(doc)
	case token.Interface:
		return p.parseInterfaceDecl(doc)
	case token.Trait:
		return p.parseTraitDecl(doc)
	default:
		if doc != nil {
			p.errorf("unexpected %v, expecting top-level statement", token.DocComment)
			return nil
		}
		return p.parseStmt()
	}
}

// ConstDecl = "const" ident "=" Expr ";" .
func (p *parser) parseConstDecl(doc *phpdoc.Block) *ConstDecl {
	cons := new(ConstDecl)
	cons.Doc = doc
	p.expect(token.Const)
	cons.Name = p.expect(token.Ident)
	p.expect(token.Assign)
	cons.X = p.parseExpr()
	p.expect(token.Semicolon)
	return cons
}

// VarDecl = var [ "=" Expr ] ";" .
func (p *parser) parseVarDecl(doc *phpdoc.Block, static bool) *VarDecl {
	v := new(VarDecl)
	v.Doc = doc
	v.Static = static
	v.Name = p.expect(token.Var)
	if p.got(token.Assign) {
		v.X = p.parseExpr()
	}
	p.expect(token.Semicolon)
	return v
}

// FuncDecl = "function" ident ParamList [ ":" Type ] BlockStmt .
func (p *parser) parseFuncDecl(doc *phpdoc.Block, static bool) *FuncDecl {
	fn := new(FuncDecl)
	fn.Doc = doc
	fn.Static = static
	p.expect(token.Function)
	fn.Name = p.expect(token.Ident)
	fn.Params = p.parseParamList()
	if p.got(token.Colon) {
		fn.Result = p.parseType()
	}
	if p.tok.Type == token.Lbrace {
		// Interfaces need not have bodies.
		fn.Body = p.parseBlockStmt()
	} else {
		p.expect(token.Semicolon)
	}
	return fn
}

// ParamList = "(" [ Param { "," Param } [ "," ] ] ")" .
// Param     = [ Type ] [ "&" ] [ "..." ] var [ "=" Lit ] .
func (p *parser) parseParamList() []*Param {
	var params []*Param
	p.expect(token.Lparen)
	for p.until(token.Rparen) {
		par := new(Param)
		par.Type = p.tryParseType()
		par.ByRef = p.got(token.And)
		par.Variadic = p.got(token.Ellipsis)
		par.Name = p.expect(token.Var)
		if p.got(token.Assign) {
			par.Default = p.parseConstExpr()
		}
		params = append(params, par)
		if p.tok.Type == token.Rparen {
			break
		}
		p.expect(token.Comma)
	}
	p.expect(token.Rparen)
	return params
}

// ClassDecl = [ "abstract" ] "class" ident [ "extends" Name ]
//             [ "implements" Name { "," Name } ]
//             "{" { UseStmt } { ClassMember } "}" .
func (p *parser) parseClassDecl(doc *phpdoc.Block) *ClassDecl {
	return p.parseClassDeclaration(doc, false)
}

// AnonymClassDecl = "class" [ "extends" Name ]
//                   [ "implements" Name { "," Name } ]
//                   "{" { UseStmt } { ClassMember } "}" .
func (p *parser) parseAnonymClassDecl() *ClassDecl {
	return p.parseClassDeclaration(nil, true)
}

func (p *parser) parseClassDeclaration(doc *phpdoc.Block, anonymous bool) *ClassDecl {
	class := new(ClassDecl)
	class.Doc = doc
	class.Abstract = p.got(token.Abstract)
	p.expect(token.Class)
	if !anonymous {
		class.Name = p.expect(token.Ident)
	}
	if p.got(token.Extends) {
		class.Extends = p.parseName()
	}
	if p.got(token.Implements) {
		for {
			class.Implements = append(class.Implements, p.parseName())
			if !p.got(token.Comma) {
				break
			}
		}
	}
	p.expect(token.Lbrace)
	for p.tok.Type == token.Use {
		class.Traits = append(class.Traits, p.parseUseStmt())
	}
	for p.until(token.Rbrace) {
		m := p.parseClassMember()
		class.Members = append(class.Members, m)
	}
	p.expect(token.Rbrace)
	return class
}

// InterfaceDecl = "interface" [ "extends" Name ] "{" { ClassMember } "}" .
func (p *parser) parseInterfaceDecl(doc *phpdoc.Block) *InterfaceDecl {
	// TODO: Really ClassMember?
	iface := new(InterfaceDecl)
	iface.Doc = doc
	p.expect(token.Interface)
	iface.Name = p.expect(token.Ident)
	if p.got(token.Extends) {
		iface.Extends = p.parseName()
	}
	p.expect(token.Lbrace)
	for p.until(token.Rbrace) {
		m := p.parseClassMember()
		iface.Members = append(iface.Members, m)
	}
	p.expect(token.Rbrace)
	return iface
}

// TraitDecl = "trait" [ "extends" Name ] "{" { ClassMember } "}" .
func (p *parser) parseTraitDecl(doc *phpdoc.Block) *TraitDecl {
	trait := new(TraitDecl)
	trait.Doc = doc
	p.expect(token.Trait)
	trait.Name = p.expect(token.Ident)
	p.expect(token.Lbrace)
	for p.until(token.Rbrace) {
		m := p.parseClassMember()
		trait.Members = append(trait.Members, m)
	}
	p.expect(token.Rbrace)
	return trait
}

// ClassMember = [ PHPDoc ] [ Visibility ]
//               ( ConstDecl | [ "static" ] VarDecl | [ "static" ] FuncDecl ) .
func (p *parser) parseClassMember() *ClassMember {
	m := new(ClassMember)
	m.Doc = p.parsePHPDoc()
	m.Vis = p.parseVisibility()
	static := p.got(token.Static)
	switch p.tok.Type {
	default:
		p.errorf("unexpected %v, expecting %v or %v", p.tok, token.Const, token.Function)
		return nil
	case token.Const:
		if static {
			p.errorf("unexpected %v in constant declaration", token.Static)
		}
		m.Decl = p.parseConstDecl(nil)
	case token.Var:
		m.Decl = p.parseVarDecl(nil, static)
	case token.Function:
		m.Decl = p.parseFuncDecl(nil, static)
	}
	return m
}

// Visibility = "public" | "protected" | "private" .
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

// PHPDoc = docComment .
func (p *parser) parsePHPDoc() *phpdoc.Block {
	if p.tok.Type != token.DocComment {
		return nil
	}
	doc, err := phpdoc.Parse(strings.NewReader(p.tok.Text))
	if err != nil && p.err == nil {
		if se, ok := err.(*phpdoc.SyntaxError); ok {
			p.err = phpDocErr(p.tok.Pos, se)
		} else {
			p.errorf("parsing PHPDoc: %v", err)
		}
	}
	p.next()
	return doc
}

func phpDocErr(p token.Pos, d *phpdoc.SyntaxError) error {
	e := &SyntaxError{p.Line, p.Column, fmt.Errorf("parsing PHPDoc: %v", d.Err)}
	if d.Line == 1 {
		e.Column += d.Column - 1
	} else {
		e.Line += d.Line - 1
		e.Column = d.Column
	}
	return e
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
	switch p.tok.Type {
	case token.Lbrace:
		return p.parseBlockStmt()
	case token.For:
		return p.parseForStmt()
	default:
		return p.parseUnknownStmt()
	}
}

// ForStmt = "for" "(" [ Expr ] ";" [ Expr ]  ";" [ Expr ]  ) Stmt .
func (p *parser) parseForStmt() Stmt {
	f := new(ForStmt)
	p.expect(token.For)
	p.expect(token.Lparen)
	if !p.got(token.Semicolon) {
		f.Init = p.parseExpr()
		p.expect(token.Semicolon)
	}
	if !p.got(token.Semicolon) {
		f.Cond = p.parseExpr()
		p.expect(token.Semicolon)
	}
	if !p.got(token.Rparen) {
		f.Post = p.parseExpr()
		p.expect(token.Rparen)
	}
	f.Body = p.parseStmt()
	return f
}

// Type = [ "?" ] Name .
func (p *parser) parseType() *Type {
	typ := p.tryParseType()
	if typ == nil {
		p.errorf("unexpected %v, expecting type", p.tok.Type)
	}
	return typ
}

func (p *parser) tryParseType() *Type {
	typ := new(Type)
	switch p.tok.Type {
	default:
		return nil
	case token.Qmark:
		typ.Nullable = true
		p.next()
		fallthrough
	case token.Ident, token.Backslash:
		typ.Name = p.parseName()
	}
	return typ
}

// Name = [ "\\" ] ident { "\\" ident } .
func (p *parser) parseName() *Name {
	id := new(Name)
	if p.got(token.Backslash) {
		id.Global = true
	}
	for {
		id.Parts = append(id.Parts, p.expect(token.Ident))
		if !p.got(token.Backslash) {
			break
		}
	}
	return id
}

// UnknownStmt = Expr ( ";" | BlockStmt ) .
func (p *parser) parseUnknownStmt() *UnknownStmt {
	stmt := new(UnknownStmt)
	stmt.X = p.parseUnknownExpr()
	switch p.tok.Type {
	case token.Semicolon:
		p.next()
	case token.Lbrace:
		stmt.Body = p.parseBlockStmt()
	}
	return stmt
}

// Expr = UnknownExpr .
func (p *parser) parseExpr() Expr { return p.parseUnknownExpr() }

// ConstExpr = BasicLit | ArrayLit .
// ArrayLit  = "[" [ ConstExpr { "," ConstExpr } [ "," ] ] "]" .
func (p *parser) parseConstExpr() Expr {
	if p.got(token.Lbrack) {
		a := new(ArrayLit)
		for !p.got(token.Rbrack) && !p.got(token.EOF) {
			a.Elems = append(a.Elems, p.parseConstExpr())
			if p.got(token.Rbrack) {
				break
			}
			p.expect(token.Comma)
		}
		return a
	}
	if p.tok.Type == token.Ident {
		// TODO: This needs some rethinking.
		n := p.parseName()
		if p.got(token.DoubleColon) {
			x := &StaticSelectorExpr{X: n}
			x.Sel = p.expect(token.Ident)
			return x
		}
		return n
	}
	return p.parseBasicLit()
}

// FuncLit      = "function" ParamList [ FuncLitScope ] [ ":" Type ] BlockStmt .
// FuncLitScope = "use" ParamList .
func (p *parser) parseFuncLit() *FuncLit {
	fn := new(FuncLit)
	p.expect(token.Function)
	fn.Params = p.parseParamList()
	if p.got(token.Use) {
		// TODO: This is easier, but enables invalid
		// syntax.
		fn.Scope = p.parseParamList()
	}
	if p.got(token.Colon) {
		fn.Result = p.parseType()
	}
	fn.Body = p.parseBlockStmt()
	return fn
}

// BasicLit = string | int | ident .
func (p *parser) parseBasicLit() Expr {
	// TODO: This needs some more work.
	switch p.tok.Type {
	default:
		p.errorf("unexpected %v, expecting lit", p.tok.Type)
		return nil
	case token.String, token.Int, token.Ident:
		lit := &UnknownExpr{[]interface{}{p.tok}}
		p.next()
		return lit
	}
}

// UnknownExpr =  ExprElem { ExprElem } .
// ExprElem    =  /* any token */ | "{" Expr "}" | FuncLit .
func (p *parser) parseUnknownExpr() *UnknownExpr {
	x := new(UnknownExpr)
	for {
		switch p.tok.Type {
		// TODO: EOF or ?>
		case token.EOF:
			p.errorf("unexpected %v, expecting %v, %v, %v or %v", p.tok, token.Semicolon, token.Lbrace, token.Rbrace, token.Rparen)
			return nil
		case token.Semicolon, token.Lbrace, token.Rbrace, token.Rparen:
			if len(x.Elems) == 0 {
				p.errorf("unexpected empty expression")
			}
			return x
		case token.Arrow:
			x.Elems = append(x.Elems, p.tok)
			p.next()
			tok := p.tok
			// Take any token that comes. Apparently you can
			// call a method that has a keyword as a name (e.g.
			// (expr)->class(args)).
			x.Elems = append(x.Elems, p.tok)
			p.next()
			if tok.Type == token.Lbrace {
				x.Elems = append(x.Elems, p.parseExpr(), p.expect(token.Rbrace))
			}
		case token.DoubleColon:
			// The next token might be "class", so we want
			// to consume 2 tokens (ignoring whitespace).
			x.Elems = append(x.Elems, p.tok)
			p.next()
			x.Elems = append(x.Elems, p.tok)
			p.next()
		case token.Lparen:
			x.Elems = append(x.Elems, p.tok)
			p.next0()
			if p.tok.Type == token.Rparen {
				// TODO: Remove special case for empty ()
				x.Elems = append(x.Elems, p.tok)
				p.next0()
				continue
			}
			x.Elems = append(x.Elems, p.parseExpr())
			x.Elems = append(x.Elems, p.expect(token.Rparen))
		case token.Class:
			x.Elems = append(x.Elems, p.parseAnonymClassDecl())
		case token.Function:
			x.Elems = append(x.Elems, p.parseFuncLit())
		default:
			x.Elems = append(x.Elems, p.tok)
			p.next0()
		}
	}
}
