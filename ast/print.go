package ast

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"mibk.io/php/token"
	"mibk.io/phpdoc"
)

// Fprint "pretty-prints" an AST node to w.
func Fprint(w io.Writer, node interface{}) error {
	w = &trimmer{output: w}
	tw := tabwriter.NewWriter(w, 0, 8, 1, '\t', tabwriter.StripEscape)
	buf := bufio.NewWriter(tw)
	p := &printer{buf: buf}
	p.print(node)
	if p.err != nil {
		return p.err
	}
	if err := p.buf.Flush(); err != nil {
		return err
	}
	return tw.Flush()
}

type indentation int

type printer struct {
	buf *bufio.Writer
	err error // sticky

	indent indentation
}

type whitespace byte

const (
	nextcol whitespace = '\v'
	tabesc  whitespace = tabwriter.Escape
	newline whitespace = '\n'
)

func (p *printer) print(args ...interface{}) {
	for _, arg := range args {
		if p.err != nil {
			return
		}

		switch arg := arg.(type) {
		case *File:
			p.print(token.OpenTag)
			switch len(arg.Pragmas) {
			default:
				p.print(newline)
				fallthrough
			case 0:
				p.print(newline)
			case 1:
				p.print(' ')
			}
			for _, d := range arg.Pragmas {
				p.print(d)
			}
			if ns := arg.Namespace; ns != nil {
				ns.Global = false // namespaces are global implicitly
				p.print(newline, token.Namespace, ' ', ns, token.Semicolon, newline)
			}
			if len(arg.UseStmts) > 0 {
				p.print(newline)
				for _, stmt := range arg.UseStmts {
					p.print(stmt, newline)
				}
			}
			for _, decl := range arg.Decls {
				p.print(newline, decl.doc(), p.indent, decl)
			}
		case *Pragma:
			p.print(token.Declare, token.Lparen)
			p.print(arg.Name, token.Assign, arg.Value)
			p.print(token.Rparen, token.Semicolon, newline)
		case *UseStmt:
			stmt := arg.Name
			stmt.Global = false // use statements are global implicitly
			p.print(token.Use, ' ', stmt, token.Semicolon)
		case *ConstDecl:
			p.print(token.Const, ' ', arg.Name, ' ', token.Assign, ' ')
			p.print(arg.X, token.Semicolon, newline)
		case *VarDecl:
			p.print(arg.Name)
			if arg.X != nil {
				p.print(' ', token.Assign, ' ', arg.X)
			}
			p.print(token.Semicolon, newline)
		case *FuncDecl:
			p.print(token.Function, ' ', arg.Name, token.Lparen, arg.Params, token.Rparen)
			if arg.Result != nil {
				p.print(token.Colon, ' ', arg.Result)
			}
			if arg.Body != nil {
				p.print(' ', arg.Body)
			} else {
				p.print(token.Semicolon)
			}
			p.print(newline)
		case []*Param:
			for i, par := range arg {
				if i > 0 {
					p.print(token.Comma, ' ')
				}
				if par.Type != nil {
					p.print(par.Type, ' ')
				}
				if par.ByRef {
					p.print(token.And)
				}
				if par.Variadic {
					p.print(token.Ellipsis)
				}
				p.print(par.Name)
				if par.Default != nil {
					p.print(' ', token.Assign, ' ', par.Default)
				}
			}
		case *ClassDecl:
			p.print(token.Class, ' ', arg.Name)
			if arg.Extends != nil {
				p.print(' ', token.Extends, ' ', arg.Extends)
			}
			if len(arg.Implements) > 0 {
				p.print(' ', token.Implements, ' ', arg.Implements[0])
				for _, n := range arg.Implements[1:] {
					p.print(token.Comma, ' ', n)
				}
			}
			p.print(newline, token.Lbrace, newline)
			for _, t := range arg.Traits {
				p.print(p.indent, t, newline)
			}
			if len(arg.Traits) > 0 {
				p.print(newline)
			}
			for i, m := range arg.Members {
				if i > 0 {
					p.print(newline)
				}
				p.print(m.Doc, p.indent, m)
			}
			p.print(p.indent-1, token.Rbrace, newline)
		case *InterfaceDecl:
			p.print(token.Interface, ' ', arg.Name)
			if arg.Extends != nil {
				p.print(' ', token.Extends, ' ', arg.Extends)
			}
			p.print(newline, token.Lbrace, newline)
			// TODO: Dedup printing members?
			for i, m := range arg.Members {
				if i > 0 {
					p.print(newline)
				}
				p.print(m.Doc, p.indent, m)
			}
			p.print(p.indent-1, token.Rbrace, newline)
		case *TraitDecl:
			p.print(token.Trait, ' ', arg.Name)
			p.print(newline, token.Lbrace, newline)
			for i, m := range arg.Members {
				if i > 0 {
					p.print(newline)
				}
				p.print(m.Doc, p.indent, m)
			}
			p.print(p.indent-1, token.Rbrace, newline)
		case *ClassMember:
			p.print(arg.Vis, arg.Decl)
		case Vis:
			switch arg {
			case Public:
				p.print(token.Public, ' ')
			case Protected:
				p.print(token.Protected, ' ')
			case Private:
				p.print(token.Private, ' ')
			case DefaultVis:
				// Don't print.
			default:
				p.err = fmt.Errorf("unknown visibility: %v", arg)
			}
		case *BlockStmt:
			p.print(token.Lbrace, newline)
			for _, stmt := range arg.List {
				p.print(p.indent, stmt, newline)
			}
			p.print(p.indent-1, token.Rbrace)
		case *UnknownStmt:
			p.print(arg.X)
			if arg.Body != nil {
				p.print(arg.Body)
			} else {
				p.print(token.Semicolon)
			}
		case *UnknownExpr:
			for i, elem := range arg.Elems {
				switch elem := elem.(type) {
				case token.Token:
					if i < len(arg.Elems)-1 || elem.Type != token.Whitespace {
						p.print(elem.Text)
					}
				default:
					p.print(elem)
				}
			}
		case *Type:
			if arg.Nullable {
				p.print(token.Qmark)
			}
			p.print(arg.Name)
		case *Name:
			for i, part := range arg.Parts {
				if i > 0 || arg.Global {
					p.print(token.Backslash)
				}
				p.print(part)
			}
		case *phpdoc.Block:
			if arg == nil {
				continue
			}
			doc := new(phpdoc.Block)
			*doc = *arg
			doc.Indent = strings.Repeat("\t", int(p.indent))
			p.err = phpdoc.Fprint(p.buf, doc)
		case token.Type:
			switch arg {
			case token.Lbrace:
				p.indent++
			case token.Rbrace:
				p.indent--
			}
			_, p.err = p.buf.WriteString(arg.String())
		case string:
			_, p.err = p.buf.WriteString(arg)
		case rune:
			_, p.err = p.buf.WriteRune(arg)
		case indentation:
			for i := 0; i < int(arg); i++ {
				p.buf.WriteByte('\t')
			}
		case whitespace:
			p.err = p.buf.WriteByte(byte(arg))
		default:
			p.err = fmt.Errorf("unsupported type %T", arg)
		}
	}
}

// The following is taken from https://golang.org/src/go/printer/printer.go.
//
// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// A trimmer is an io.Writer filter for stripping tabwriter.Escape
// characters, trailing blanks and tabs, and for converting formfeed
// and vtab characters into newlines and htabs (in case no tabwriter
// is used). Text bracketed by tabwriter.Escape characters is passed
// through unchanged.
//
type trimmer struct {
	output io.Writer
	state  int
	space  []byte
}

// trimmer is implemented as a state machine.
// It can be in one of the following states:
const (
	inSpace  = iota // inside space
	inEscape        // inside text bracketed by tabwriter.Escapes
	inText          // inside text
)

func (p *trimmer) resetSpace() {
	p.state = inSpace
	p.space = p.space[0:0]
}

var aNewline = []byte("\n")

func (p *trimmer) Write(data []byte) (n int, err error) {
	// invariants:
	// p.state == inSpace:
	//	p.space is unwritten
	// p.state == inEscape, inText:
	//	data[m:n] is unwritten
	m := 0
	var b byte
	for n, b = range data {
		if b == '\v' {
			b = '\t' // convert to htab
		}
		switch p.state {
		case inSpace:
			switch b {
			case '\t', ' ':
				p.space = append(p.space, b)
			case '\n', '\f':
				p.resetSpace() // discard trailing space
				_, err = p.output.Write(aNewline)
			case tabwriter.Escape:
				_, err = p.output.Write(p.space)
				p.state = inEscape
				m = n + 1 // +1: skip tabwriter.Escape
			default:
				_, err = p.output.Write(p.space)
				p.state = inText
				m = n
			}
		case inEscape:
			if b == tabwriter.Escape {
				_, err = p.output.Write(data[m:n])
				p.resetSpace()
			}
		case inText:
			switch b {
			case '\t', ' ':
				_, err = p.output.Write(data[m:n])
				p.resetSpace()
				p.space = append(p.space, b)
			case '\n', '\f':
				_, err = p.output.Write(data[m:n])
				p.resetSpace()
				if err == nil {
					_, err = p.output.Write(aNewline)
				}
			case tabwriter.Escape:
				_, err = p.output.Write(data[m:n])
				p.state = inEscape
				m = n + 1 // +1: skip tabwriter.Escape
			}
		default:
			panic("unreachable")
		}
		if err != nil {
			return
		}
	}
	n = len(data)

	switch p.state {
	case inEscape, inText:
		_, err = p.output.Write(data[m:n])
		p.resetSpace()
	}

	return
}
