package token_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"mibk.io/php/token"
)

func pos(posStr string) token.Pos {
	var pos token.Pos
	fmt.Sscanf(posStr, "%d:%d", &pos.Line, &pos.Column)
	return pos
}

func TestScanner(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []token.Token
	}{{
		"only HTML",
		`doesn't
actually have to be a <html>`,
		[]token.Token{
			{token.InlineHTML, "doesn't\nactually have to be a <html>", pos("1:1")},
			{token.EOF, "", pos("2:28")},
		},
	}, {
		"tease opening",
		`< <?ph  <?p <?hp nic <?php`,
		[]token.Token{
			{token.InlineHTML, "< <?ph  <?p <?hp nic ", pos("1:1")},
			{token.OpenTag, "<?php", pos("1:22")},
			{token.EOF, "", pos("1:27")},
		},
	}, {
		"basic PHP",
		`<html> <?php
echo 'ahoj';`,
		[]token.Token{
			{token.InlineHTML, "<html> ", pos("1:1")},
			{token.OpenTag, "<?php", pos("1:8")},
			{token.Whitespace, "\n", pos("1:13")},
			{token.Ident, "echo", pos("2:1")},
			{token.Whitespace, " ", pos("2:5")},
			{token.String, `'ahoj'`, pos("2:6")},
			{token.Semicolon, ";", pos("2:12")},
			{token.EOF, "", pos("2:13")},
		},
	}, {
		"comments",
		`<?php // line comment
namespace /*block */ DateTime/** comments*/;`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.Comment, "// line comment", pos("1:7")},
			{token.Whitespace, "\n", pos("1:22")},
			{token.Ident, "namespace", pos("2:1")},
			{token.Whitespace, " ", pos("2:10")},
			{token.Comment, "/*block */", pos("2:11")},
			{token.Whitespace, " ", pos("2:21")},
			{token.Ident, "DateTime", pos("2:22")},
			{token.Comment, "/** comments*/", pos("2:30")},
			{token.Semicolon, ";", pos("2:44")},
			{token.EOF, "", pos("2:45")},
		},
	}, {
		"single quoted strings",
		`<?php'\'\\' '\\' '\'' '\\n\\\'''
\''`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.String, `'\'\\'`, pos("1:6")},
			{token.Whitespace, " ", pos("1:12")},
			{token.String, `'\\'`, pos("1:13")},
			{token.Whitespace, " ", pos("1:17")},
			{token.String, `'\''`, pos("1:18")},
			{token.Whitespace, " ", pos("1:22")},
			{token.String, `'\\n\\\''`, pos("1:23")},
			{token.String, "'\n\\''", pos("1:32")},
			{token.EOF, "", pos("2:4")},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := token.NewScanner(strings.NewReader(tt.input))

			var got []token.Token
			for {
				tok := sc.Next()
				got = append(got, tok)
				if tok.Type == token.EOF {
					break
				}
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("tokens don't match: (-got +want)\n%s", diff)
			}
		})
	}
}
