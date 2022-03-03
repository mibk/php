package token

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type ScanError struct {
	Pos Pos
	Err error
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("line:%v: %v", e.Pos, e.Err)
}

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
	switch {
	case t.Type == EOF,
		symbolStart < t.Type && t.Type < symbolEnd,
		keywordStart < t.Type && t.Type < keywordEnd:
		return t.Type.String()
	default:
		return fmt.Sprintf("%v(%q)", t.Type, t.Text)
	}
}

//go:generate stringer -type Type -linecomment

type Type uint

const (
	Illegal Type = iota
	EOF
	Whitespace
	Comment
	DocComment

	Ident
	Int
	Float
	String
	Var
	InlineHTML

	symbolStart
	OpenTag     // <?php
	CloseTag    // ?>
	Dollar      // $
	Backslash   // \
	Qmark       // ?
	Lparen      // (
	Rparen      // )
	Lbrack      // [
	Rbrack      // ]
	Lbrace      // {
	Rbrace      // }
	Add         // +
	Sub         // -
	Assign      // =
	Lt          // <
	Gt          // >
	Period      // .
	Comma       // ,
	Colon       // :
	DoubleColon // ::
	Semicolon   // ;
	Ellipsis    // ...
	Or          // |
	And         // &
	Quo         // /
	Shl         // <<
	Shr         // >>
	Arrow       // ->
	DoubleArrow // =>
	symbolEnd

	keywordStart
	Abstract   // abstract
	As         // as
	Break      // break
	Case       // case
	Catch      // catch
	Class      // class
	Clone      // clone
	Const      // const
	Continue   // continue
	Declare    // declare
	Default    // default
	Do         // do
	Else       // else
	Extends    // extends
	Final      // final
	Finally    // finally
	Fn         // fn
	For        // for
	Foreach    // foreach
	Function   // function
	Goto       // goto
	If         // if
	Implements // implements
	Instanceof // instanceof
	Insteadof  // insteadof
	Interface  // interface
	Namespace  // namespace
	New        // new
	Private    // private
	Protected  // protected
	Public     // public
	Return     // return
	Static     // static
	Switch     // switch
	Throw      // throw
	Trait      // trait
	Try        // try
	Use        // use
	While      // while
	keywordEnd
)

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for typ := keywordStart + 1; typ < keywordEnd; typ++ {
		s := typ.String()
		keywords[s] = Token{Type: typ, Text: s}
	}
}

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
	err   error

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

func (s *Scanner) Next() (tok Token) {
	defer func() {
		switch tok.Type {
		case OpenTag:
			s.state = inPHP
		case CloseTag:
			s.state = inHTML
		}
	}()

	if len(s.queue) > 0 {
		tok, s.queue = s.queue[0], s.queue[1:]
		return tok
	}

	pos := s.pos()
	switch s.state {
	default:
		panic(fmt.Sprintf("unknown state: %d", s.state))
	case inHTML:
		tok = s.scanInlineHTML()
	case inPHP:
		tok = s.scanAny()
		if typ := tok.Type; symbolStart < typ && typ < symbolEnd {
			tok.Text = typ.String()
		}
	}
	tok.Pos = pos
	return tok
}

func (s *Scanner) Err() error { return s.err }

func (s *Scanner) errorf(format string, args ...interface{}) Token {
	if s.err == nil {
		s.err = &ScanError{s.pos(), fmt.Errorf(format, args...)}
	}
	return Token{Type: EOF}
}

func (s *Scanner) pos() Pos { return Pos{Line: s.line, Column: s.col} }

func (s *Scanner) read() rune {
	if s.done {
		return eof
	}
	r, _, err := s.r.ReadRune()
	if err != nil {
		if err != io.EOF {
			s.err = err
		}
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
		if s.peek() == '>' {
			s.read()
			return Token{Type: CloseTag}
		}
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
	case '=':
		if s.peek() == '>' {
			s.read()
			return Token{Type: DoubleArrow}
		}
		return Token{Type: Assign}
	case '+':
		return Token{Type: Add}
	case '-':
		if s.peek() == '>' {
			s.read()
			return Token{Type: Arrow}
		}
		return Token{Type: Sub}
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
	case '.':
		switch r2 := s.peek(); {
		case r2 == r:
			s.read()
			if s.peek() != r {
				return Token{Type: Illegal, Text: ".."}
			}
			s.read()
			return Token{Type: Ellipsis}
		case isDigit(r2):
			b := new(strings.Builder)
			b.WriteRune(r)
			return s.scanFloat(b)
		default:
			return Token{Type: Period}
		}
	case ',':
		return Token{Type: Comma}
	case ':':
		if s.peek() == r {
			s.read()
			return Token{Type: DoubleColon}
		}
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
		if isDigit(r) {
			return s.scanNumber(r)
		}
		s.unread()
		if id := s.scanIdent(); id != "" {
			if tok, ok := keywords[id]; ok {
				return tok
			}
			if id == "elseif" {
				// Ugly special case.
				t := keywords["if"]
				t.Pos = s.pos()
				t.Pos.Column -= 2
				s.queue = append(s.queue, t)
				return keywords["else"]
			}
			return Token{Type: Ident, Text: id}
		}
		s.read()
		return Token{Type: Illegal, Text: string(r)}
	}
}

func (s *Scanner) scanInlineHTML() Token {
	const openTag = "<?php"
	var i int
	var canEnd bool
	var b strings.Builder
	for {
		switch r := s.read(); r {
		case rune(openTag[i]):
			i++
			if i == len(openTag) {
				canEnd = true
				i = 0
			}
		case ' ', '\t', '\r', '\n':
			if canEnd {
				s.unread()
				tok := Token{Type: OpenTag, Text: openTag}
				if b.Len() > 0 {
					tok.Pos.Line, tok.Pos.Column = s.line, s.col-len(openTag)
					s.queue = append(s.queue, tok)
					tok = Token{Type: InlineHTML, Text: b.String()}
				}
				return tok
			}
			fallthrough
		default:
			if canEnd {
				i = len(openTag)
			}
			canEnd = false
			b.WriteString(openTag[:i])
			if r == eof {
				if b.Len() == 0 {
					return Token{Type: EOF}
				}
				s.unread()
				return Token{Type: InlineHTML, Text: b.String()}
			}
			i = 0
			b.WriteRune(r)
		}
	}
}

func (s *Scanner) scanLineComment(start string) Token {
	var b strings.Builder
	for {
		switch r := s.read(); r {
		case '?':
			// Close tags end line comments, too.
			if s.peek() == '>' {
				s.read()
				tok := Token{Type: CloseTag, Text: "?>"}
				tok.Pos.Line, tok.Pos.Column = s.line, s.col-2
				s.queue = append(s.queue, tok)
				return Token{Type: Comment, Text: start + b.String()}
			}
			fallthrough
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
			tok := Token{Type: Comment, Text: "/*" + b.String() + "*/"}
			if strings.HasPrefix(tok.Text, "/**") {
				tok.Type = DocComment
			}
			return tok
		case r == eof:
			return s.errorf("unterminated block comment")
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
			// It would be nice if PHP disallowed unknown escape
			// sequences. Was tempted to disallow it here but then
			// realized it would be unnecessary painful to write
			// regexes, e.g.:
			//
			//	'\\d+.\\d{1,2}'
			//
			// Be compatible with PHP for now.
			b.WriteRune(s.read())
		case '\'':
			return Token{Type: String, Text: "'" + b.String()}
		case eof:
			return s.errorf("string not terminated")
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
			// Allow all escape sequences, even unknown ones.
			// Be compatible with PHP for now.
			b.WriteRune(s.read())
		case '"':
			return Token{Type: String, Text: `"` + b.String()}
		case eof:
			return s.errorf("string not terminated")
		}
	}
}

func (s *Scanner) scanHereDoc() Token {
	var b strings.Builder
	ws := s.scanWhitespace()
	if strings.ContainsAny(ws.Text, "\r\n") || s.peek() == eof {
		// TODO: err position might be wrong.
		return s.errorf("missing opening heredoc identifier")
	}
	b.WriteString(ws.Text)
	var quote rune
	switch r := s.peek(); r {
	case '"', '\'':
		s.read()
		b.WriteRune(r)
		quote = r
	}
	delim := s.scanIdent()
	if delim == "" {
		return s.errorf("invalid opening heredoc identifier")
	}
	b.WriteString(delim)
	if quote != 0 {
		if s.read() != quote {
			// TODO: Different message for nowdoc?
			return s.errorf("quoted heredoc identifier not terminated")
		}
		b.WriteRune(quote)
	}
	ws = s.scanWhitespace()
	if !strings.ContainsRune(ws.Text, '\n') {
		return s.errorf("no newline after identifier in heredoc")
	}
	b.WriteString(ws.Text)
	for {
		// TODO: Check escape characters for heredoc.
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
			if sep := s.peek(); sep == ';' || sep == ',' {
				// There might be a semicolon or comma after heredoc closing identifier.
				t := Token{Type: Semicolon, Text: string(sep), Pos: s.pos()}
				if sep == ',' {
					t.Type = Comma
				}
				toks = append(toks, t)
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
			return s.errorf("heredoc not terminated")
		}
	}
}

func (s *Scanner) scanNumber(r rune) Token {
	// TODO: Add support for _
	if r == '0' {
		switch r := s.peek(); {
		case isDigit(r):
			return s.scanOctal()
		case r == 'x' || r == 'X':
			return s.scanHexa(s.read())
		case r == 'b' || r == 'B':
			return s.scanBinary(s.read())
		}
	}
	b := new(strings.Builder)
	b.WriteRune(r)
	s.scanDecimal(b)
	if s.peek() == '.' {
		return s.scanFloat(b)
	}
	return Token{Type: Int, Text: b.String()}
}

func (s *Scanner) scanDecimal(b *strings.Builder) {
	for isDigit(s.peek()) {
		b.WriteRune(s.read())
	}
}

func (s *Scanner) scanOctal() Token {
	var b strings.Builder
	for {
		switch r := s.peek(); r {
		default:
			return Token{Type: Int, Text: "0" + b.String()}
		case '8', '9':
			return s.errorf("invalid digit %c in octal literal", r)
		case '0', '1', '2', '3', '4', '5', '6', '7':
			b.WriteRune(s.read())
		}
	}
}

func (s *Scanner) scanHexa(delim rune) Token {
	var b strings.Builder
	for {
		switch r := s.peek(); {
		default:
			return Token{Type: Int, Text: "0" + string(delim) + b.String()}
		case isDigit(r) || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F':
			b.WriteRune(s.read())
		}
	}
}

func (s *Scanner) scanBinary(delim rune) Token {
	var b strings.Builder
	for {
		switch r := s.peek(); r {
		default:
			return Token{Type: Int, Text: "0" + string(delim) + b.String()}
		case '0', '1':
			b.WriteRune(s.read())
		}
	}
}

func (s *Scanner) scanFloat(b *strings.Builder) Token {
	b.WriteRune(s.read())
	s.scanDecimal(b)
	if r := s.peek(); r == 'e' || r == 'E' {
		b.WriteRune(s.read())
		if r := s.peek(); r == '+' || r == '-' {
			b.WriteRune(s.read())
		}
		s.scanDecimal(b)
	}
	return Token{Type: Float, Text: b.String()}
}

func isDigit(r rune) bool { return '0' <= r && r <= '9' }
