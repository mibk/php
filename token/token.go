package token

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type Pos struct {
	Line, Column int
}

func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token struct {
	Type Type
	Text string
	Pos  Pos
}

func (t Token) String() string {
	if symbolStart < t.Type && t.Type < symbolEnd || t.Type == EOF {
		return t.Type.String()
	}
	return fmt.Sprintf("%v(%q)", t.Type, t.Text)
}

//go:generate stringer -type Type -linecomment

type Type int

const (
	Illegal Type = iota
	EOF
	Whitespace
	Comment

	Ident
	String
	Var
	InlineHTML

	symbolStart
	OpenTag   // <?php
	Dollar    // $
	Backslash // \
	Qmark     // ?
	Lparen    // (
	Rparen    // )
	Lbrack    // [
	Rbrack    // ]
	Lbrace    // {
	Rbrace    // }
	Lt        // <
	Gt        // >
	Comma     // ,
	Colon     // :
	Semicolon // ;
	Ellipsis  // ...
	Or        // |
	And       // &
	Quo       // /
	Shl       // <<
	Shr       // >>
	symbolEnd
)

const eof = -1

const (
	inHTML = iota
	inPHP
)

type Scanner struct {
	r     *bufio.Reader
	state uint
	queue []Token
	done  bool

	line, col   int
	lastLineLen int
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r:    bufio.NewReader(r),
		line: 1,
		col:  1,
	}
}

const openTag = "<?php"

func (s *Scanner) Next() Token {
	if len(s.queue) > 0 {
		tok := s.queue[0]
		s.queue = s.queue[1:]
		return tok
	}

	pos := s.pos()
	var tok Token
	switch s.state {
	default:
		panic(fmt.Sprintf("unknown state: %d", s.state))
	case inHTML:
		tok, s.state = s.scanInlineHTML()
	case inPHP:
		tok = s.scanAny()
		if typ := tok.Type; symbolStart < typ && typ < symbolEnd {
			tok.Text = typ.String()
		}
	}
	tok.Pos = pos
	return tok
}

func (s *Scanner) pos() Pos { return Pos{Line: s.line, Column: s.col} }

func (s *Scanner) read() rune {
	if s.done {
		return eof
	}
	r, _, err := s.r.ReadRune()
	if err != nil {
		s.done = true
		return eof
	}
	if r == '\n' {
		s.line++
		s.lastLineLen, s.col = s.col, 1
	} else {
		s.col++
	}
	return r
}

func (s *Scanner) unread() {
	if s.done {
		return
	}
	if err := s.r.UnreadRune(); err != nil {
		// UnreadRune returns an error only on invalid use.
		panic(err)
	}
	s.col--
	if s.col == 0 {
		s.col = s.lastLineLen
		s.line--
	}
}

func (s *Scanner) peek() rune {
	r := s.read()
	s.unread()
	return r
}

func (s *Scanner) scanAny() (tok Token) {
	switch r := s.read(); r {
	case eof:
		return Token{Type: EOF}
	case '/':
		switch s.read() {
		case '/':
			return s.scanLineComment("//")
		case '*':
			return s.scanBlockComment()
		default:
			s.unread()
			return Token{Type: Quo}
		}
	case '#':
		return s.scanLineComment("#")
	case '$':
		if id := s.scanIdent(); id != "" {
			return Token{Type: Var, Text: "$" + id}
		}
		return Token{Type: Dollar}
	case '\\':
		return Token{Type: Backslash}
	case '?':
		return Token{Type: Qmark}
	case '(':
		return Token{Type: Lparen}
	case ')':
		return Token{Type: Rparen}
	case '[':
		return Token{Type: Lbrack}
	case ']':
		return Token{Type: Rbrack}
	case '{':
		return Token{Type: Lbrace}
	case '}':
		return Token{Type: Rbrace}
	case '<':
		if s.peek() == r {
			s.read()
			if s.peek() == r {
				s.read()
				return s.scanHereDoc()
			}
			return Token{Type: Shl}
		}
		return Token{Type: Lt}
	case '>':
		if s.peek() == r {
			s.read()
			return Token{Type: Shr}
		}
		return Token{Type: Gt}
	case ',':
		return Token{Type: Comma}
	case ':':
		return Token{Type: Colon}
	case ';':
		return Token{Type: Semicolon}
	case '|':
		return Token{Type: Or}
	case '&':
		return Token{Type: And}
	case ' ', '\t', '\r', '\n':
		s.unread()
		return s.scanWhitespace()
	case '\'':
		return s.scanSingleQuoted()
	case '"':
		return s.scanDoubleQuoted()
	default:
		s.unread()
		if id := s.scanIdent(); id != "" {
			return Token{Type: Ident, Text: id}
		}
		s.read()
		return Token{Type: Illegal, Text: string(r)}
	}
}

func (s *Scanner) scanInlineHTML() (Token, uint) {
	var i int
	var b strings.Builder
	for {
		switch r := s.read(); r {
		case rune(openTag[i]):
			i++
			if i == len(openTag) {
				tok := Token{Type: OpenTag, Text: openTag}
				if b.Len() > 0 {
					tok.Pos.Line, tok.Pos.Column = s.line, s.col-i
					s.queue = append(s.queue, tok)
					tok = Token{Type: InlineHTML, Text: b.String()}
				}
				return tok, inPHP
			}
		default:
			b.WriteString(openTag[:i])
			i = 0
			b.WriteRune(r)
		case eof:
			if b.Len() == 0 {
				return Token{Type: EOF}, inHTML
			}
			s.unread()
			return Token{Type: InlineHTML, Text: b.String()}, inHTML
		}
	}
}

func (s *Scanner) scanLineComment(start string) Token {
	var b strings.Builder
	for {
		switch r := s.read(); r {
		default:
			b.WriteRune(r)
		case '\n', eof:
			s.unread()
			return Token{Type: Comment, Text: start + b.String()}
		}
	}
}

func (s *Scanner) scanBlockComment() Token {
	var b strings.Builder
	for {
		switch r := s.read(); {
		default:
			b.WriteRune(r)
		case r == '*' && s.peek() == '/':
			s.read()
			return Token{Type: Comment, Text: "/*" + b.String() + "*/"}
		case r == eof:
			// TODO: don't panic
			panic("unterminated block comment")
		}
	}
}

func (s *Scanner) scanIdent() string {
	var b strings.Builder
	for {
		switch r := s.read(); {
		case r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= utf8.RuneSelf:
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			if b.Len() > 0 {
				b.WriteRune(r)
				continue
			}
			fallthrough
		default:
			s.unread()
			return b.String()
		}
	}
}

func (s *Scanner) scanWhitespace() Token {
	var b strings.Builder
	for {
		switch r := s.read(); r {
		case ' ', '\t', '\r', '\n':
			b.WriteRune(r)
		default:
			s.unread()
			return Token{Type: Whitespace, Text: b.String()}
		}
	}
}

func (s *Scanner) scanSingleQuoted() Token {
	var b strings.Builder
	for {
		r := s.read()
		b.WriteRune(r)
		switch r {
		case '\\':
			switch s.peek() {
			case '\\', '\'':
				b.WriteRune(s.read())
			default:
				// Here we differ from PHP; we don't ignore unknown
				// escape sequences.
				// TODO: don't panic
				panic("illegal escape char")
			}
		case '\'':
			return Token{Type: String, Text: "'" + b.String()}
		case eof:
			// TODO: Do not panic.
			panic("string not terminated")
		}
	}
}

func (s *Scanner) scanDoubleQuoted() Token {
	var b strings.Builder
	for {
		r := s.read()
		b.WriteRune(r)
		switch r {
		case '\\':
			switch s.peek() {
			// TODO: Add support for
			// - octal notation: \[0-7]{1,3}
			// - hex notation: \x[0-9A-Fa-f]{1,2}
			// - UTF-8 codepoint: \u{[0-9A-Fa-f]+}
			case '\\', '"', 'n', 'r', 't', 'v', 'e', 'f', '$':
				b.WriteRune(s.read())
			default:
				// Here we differ from PHP; we don't ignore unknown
				// escape sequences.
				// TODO: don't panic
				//panic("illegal escape char")
			}
		case '"':
			return Token{Type: String, Text: `"` + b.String()}
		case eof:
			// TODO: Do not panic.
			panic("string not terminated")
		}
	}
}

func (s *Scanner) scanHereDoc() Token {
	// TODO: don't panic
	var b strings.Builder
	ws := s.scanWhitespace()
	if strings.ContainsAny(ws.Text, "\r\n") {
		panic("missing opening heredoc identifier")
	}
	b.WriteString(ws.Text)
	quoted := false
	if s.peek() == '"' {
		b.WriteRune(s.read())
		quoted = true
	}
	delim := s.scanIdent()
	if delim == "" {
		panic("invalid opening identifier")
	}
	b.WriteString(delim)
	if quoted {
		b.WriteRune('"')
		if s.read() != '"' {
			panic("quoted heredoc identifier not terminated")
		}
	}
	ws = s.scanWhitespace()
	if !strings.ContainsRune(ws.Text, '\n') {
		panic("no newline after identifier in heredoc")
	}
	b.WriteString(ws.Text)
	for {
		r := s.read()
		b.WriteRune(r)
		switch r {
		case '\n':
			id := s.scanIdent()
			b.WriteString(id)
			if id != delim {
				continue
			}
			// This is a heredoc end candidate. We need to check for a newline.
			toks := make([]Token, 0, 2)
			if s.peek() == ';' {
				// There might be a semicolon after heredoc closing identifier.
				toks = append(toks, Token{Type: Semicolon, Text: ";", Pos: s.pos()})
				s.read()
			}
			if pos, ws := s.pos(), s.scanWhitespace(); ws.Text != "" {
				ws.Pos = pos
				toks = append(toks, ws)
				if strings.ContainsRune(ws.Text, '\n') {
					s.queue = append(s.queue, toks...)
					return Token{Type: String, Text: "<<<" + b.String()}
				}
			} else {
				// It wasn't a closing identifier after all.
				for _, t := range toks {
					b.WriteString(t.Text)
				}
			}
		case eof:
			// TODO: Do not panic.
			panic("heredoc not terminated")
		}
	}
}
