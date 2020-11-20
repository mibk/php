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
	VarName
	InlineHTML

	symbolStart
	OpenTag   // <?php
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
	symbolEnd
)

const eof = -1

const (
	inHTML = iota
	atOpenTag
	inPHP
	done
)

type Scanner struct {
	r     *bufio.Reader
	state uint

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

func (sc *Scanner) Next() Token {
	pos := Pos{Line: sc.line, Column: sc.col}
	var tok Token
	switch sc.state {
	case inHTML:
		tok, sc.state = sc.scanInlineHTML()
	case atOpenTag:
		tok = Token{Type: OpenTag, Text: openTag}
		pos.Column -= len(tok.Text) // was already read
		sc.state++
	case inPHP:
		tok = sc.scanAny()
		if typ := tok.Type; symbolStart < typ && typ < symbolEnd {
			tok.Text = typ.String()
		}
	}
	tok.Pos = pos
	return tok
}

func (sc *Scanner) read() rune {
	r, _, err := sc.r.ReadRune()
	if err != nil {
		return eof
	}
	if r == '\n' {
		sc.line++
		sc.lastLineLen, sc.col = sc.col, 1
	} else {
		sc.col++
	}
	return r
}

func (sc *Scanner) unread() {
	sc.r.UnreadRune()
	sc.col--
	if sc.col == 0 {
		sc.col = sc.lastLineLen
		sc.line--
	}
}

func (sc *Scanner) peek() rune {
	r := sc.read()
	sc.unread()
	return r
}

func (sc *Scanner) scanAny() (tok Token) {
	switch r := sc.read(); r {
	case eof:
		return Token{Type: EOF}
	case '/':
		switch sc.read() {
		case '/':
			return sc.scanLineComment()
		case '*':
			return sc.scanBlockComment()
		default:
			sc.unread()
			return Token{Type: Quo}
		}
	// case '$':
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
		return Token{Type: Lt}
	case '>':
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
	case ' ', '\t', '\n':
		return sc.scanWhitespace(r)
	case '\'':
		return sc.scanSingleQuoted()
	default:
		sc.unread()
		if id := sc.scanIdentName(); id != "" {
			return Token{Type: Ident, Text: id}
		}
		sc.read()
		return Token{Type: Illegal, Text: string(r)}
	}
}

func (sc *Scanner) scanInlineHTML() (Token, uint) {
	var i int
	var b strings.Builder
	for {
		switch r := sc.read(); r {
		case rune(openTag[i]):
			i++
			if i == len(openTag) {
				if b.Len() == 0 {
					return Token{Type: OpenTag, Text: openTag}, inPHP
				}
				return Token{Type: InlineHTML, Text: b.String()}, atOpenTag
			}
		default:
			b.WriteString(openTag[:i])
			i = 0
			b.WriteRune(r)
		case eof:
			if b.Len() == 0 {
				return Token{Type: EOF}, inHTML
			}
			sc.unread()
			return Token{Type: InlineHTML, Text: b.String()}, inHTML
		}
	}
}

func (sc *Scanner) scanLineComment() Token {
	var b strings.Builder
	for {
		switch r := sc.read(); r {
		default:
			b.WriteRune(r)
		case '\n', eof:
			sc.unread()
			return Token{Type: Comment, Text: "//" + b.String()}
		}
	}
}

func (sc *Scanner) scanBlockComment() Token {
	var b strings.Builder
	for {
		switch r := sc.read(); {
		default:
			b.WriteRune(r)
		case r == '*' && sc.peek() == '/':
			sc.read()
			return Token{Type: Comment, Text: "/*" + b.String() + "*/"}
		case r == eof:
			// TODO: don't panic
			panic("unterminated block comment")
		}
	}
}

func (sc *Scanner) scanIdentName() string {
	var b strings.Builder
	for {
		switch r := sc.read(); {
		case r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= utf8.RuneSelf:
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			if b.Len() > 0 {
				b.WriteRune(r)
				continue
			}
			fallthrough
		default:
			sc.unread()
			return b.String()
		}
	}
}

func (sc *Scanner) scanWhitespace(init rune) Token {
	var b strings.Builder
	b.WriteRune(init)
	for {
		switch r := sc.read(); r {
		case ' ', '\t':
			b.WriteRune(r)
		default:
			sc.unread()
			return Token{Type: Whitespace, Text: b.String()}
		}
	}
}

func (sc *Scanner) scanSingleQuoted() Token {
	var b strings.Builder
	for {
		r := sc.read()
		b.WriteRune(r)
		switch r {
		case '\\':
			switch sc.peek() {
			case '\\', '\'':
				b.WriteRune(sc.read())
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
