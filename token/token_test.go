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
			{token.EOF, "", pos("2:29")},
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
			{token.Whitespace, "\n\n   ", pos("1:13")},
			{token.Ident, "echo", pos("3:4")},
			{token.Whitespace, " ", pos("3:8")},
			{token.String, `'ahoj'`, pos("3:9")},
			{token.Semicolon, ";", pos("3:15")},
			{token.EOF, "", pos("3:16")},
		},
	}, {
		"comments",
		`<?php // line comment
namespace /*block */ DateTime/** comments*/;# another line comm.`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.Comment, "// line comment", pos("1:7")},
			{token.Whitespace, "\n", pos("1:22")},
			{token.Namespace, "namespace", pos("2:1")},
			{token.Whitespace, " ", pos("2:10")},
			{token.Comment, "/*block */", pos("2:11")},
			{token.Whitespace, " ", pos("2:21")},
			{token.Ident, "DateTime", pos("2:22")},
			{token.Comment, "/** comments*/", pos("2:30")},
			{token.Semicolon, ";", pos("2:44")},
			{token.Comment, "# another line comm.", pos("2:45")},
			{token.EOF, "", pos("2:65")},
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
	}, {
		"double quoted strings",
		`<?php"\"\\" "\\" "\"" "\\'\\\"""
\""
"\n\r\t\v\e\f\$"`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.String, `"\"\\"`, pos("1:6")},
			{token.Whitespace, " ", pos("1:12")},
			{token.String, `"\\"`, pos("1:13")},
			{token.Whitespace, " ", pos("1:17")},
			{token.String, `"\""`, pos("1:18")},
			{token.Whitespace, " ", pos("1:22")},
			{token.String, `"\\'\\\""`, pos("1:23")},
			{token.String, "\"\n\\\"\"", pos("1:32")},
			{token.Whitespace, "\n", pos("2:4")},
			{token.String, "\"\\n\\r\\t\\v\\e\\f\\$\"", pos("3:1")},
			{token.EOF, "", pos("3:17")},
		},
	}, {
		"variables",
		`<?php $žluťoučký;$$kůň;`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.Var, "$žluťoučký", pos("1:7")},
			{token.Semicolon, ";", pos("1:17")},
			{token.Dollar, "$", pos("1:18")},
			{token.Var, "$kůň", pos("1:19")},
			{token.Semicolon, ";", pos("1:23")},
			{token.EOF, "", pos("1:24")},
		},
	}, {
		"binary operators",
		`<?php<><<>>`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Lt, "<", pos("1:6")},
			{token.Gt, ">", pos("1:7")},
			{token.Shl, "<<", pos("1:8")},
			{token.Shr, ">>", pos("1:10")},
			{token.EOF, "", pos("1:12")},
		},
	}, {
		"heredoc",
		`<?php<<<	 END ` + `
buffalo
  END
END:
END;nic
END;	` + `
<<<"HERE"
there
HERE
`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.String, "<<<\t END \nbuffalo\n  END\nEND:\nEND;nic\nEND", pos("1:6")},
			{token.Semicolon, ";", pos("6:4")},
			{token.Whitespace, "\t\n", pos("6:5")},
			{token.String, "<<<\"HERE\"\nthere\nHERE", pos("7:1")},
			{token.Whitespace, "\n", pos("9:5")},
			{token.EOF, "", pos("10:1")},
		},
	}, {
		"nowdoc",
		`<?php<<<	 'NOWdoc' ` + `
weather
  NOWdoc
NOWdoc:
NOWdoc;nada
NOWdoc;	` + `
`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.String, "<<<\t 'NOWdoc' \nweather\n  NOWdoc\nNOWdoc:\nNOWdoc;nada\nNOWdoc", pos("1:6")},
			{token.Semicolon, ";", pos("6:7")},
			{token.Whitespace, "\t\n", pos("6:8")},
			{token.EOF, "", pos("7:1")},
		},
	}, {
		"keywords",
		`<?php
abstract as
break
callable case catch class clone const continue
default do
else elseif extends
final finally fn for foreach function
goto
if implements instanceof insteadof interface
namespace new
parent private protected public
return
self static switch
throw trait try
use
while
`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, "\n", pos("1:6")},
			{token.Abstract, "abstract", pos("2:1")},
			{token.Whitespace, " ", pos("2:9")},
			{token.As, "as", pos("2:10")},
			{token.Whitespace, "\n", pos("2:12")},
			{token.Break, "break", pos("3:1")},
			{token.Whitespace, "\n", pos("3:6")},
			{token.Callable, "callable", pos("4:1")},
			{token.Whitespace, " ", pos("4:9")},
			{token.Case, "case", pos("4:10")},
			{token.Whitespace, " ", pos("4:14")},
			{token.Catch, "catch", pos("4:15")},
			{token.Whitespace, " ", pos("4:20")},
			{token.Class, "class", pos("4:21")},
			{token.Whitespace, " ", pos("4:26")},
			{token.Clone, "clone", pos("4:27")},
			{token.Whitespace, " ", pos("4:32")},
			{token.Const, "const", pos("4:33")},
			{token.Whitespace, " ", pos("4:38")},
			{token.Continue, "continue", pos("4:39")},
			{token.Whitespace, "\n", pos("4:47")},
			{token.Default, "default", pos("5:1")},
			{token.Whitespace, " ", pos("5:8")},
			{token.Do, "do", pos("5:9")},
			{token.Whitespace, "\n", pos("5:11")},
			{token.Else, "else", pos("6:1")},
			{token.Whitespace, " ", pos("6:5")},
			{token.Elseif, "elseif", pos("6:6")},
			{token.Whitespace, " ", pos("6:12")},
			{token.Extends, "extends", pos("6:13")},
			{token.Whitespace, "\n", pos("6:20")},
			{token.Final, "final", pos("7:1")},
			{token.Whitespace, " ", pos("7:6")},
			{token.Finally, "finally", pos("7:7")},
			{token.Whitespace, " ", pos("7:14")},
			{token.Fn, "fn", pos("7:15")},
			{token.Whitespace, " ", pos("7:17")},
			{token.For, "for", pos("7:18")},
			{token.Whitespace, " ", pos("7:21")},
			{token.Foreach, "foreach", pos("7:22")},
			{token.Whitespace, " ", pos("7:29")},
			{token.Function, "function", pos("7:30")},
			{token.Whitespace, "\n", pos("7:38")},
			{token.Goto, "goto", pos("8:1")},
			{token.Whitespace, "\n", pos("8:5")},
			{token.If, "if", pos("9:1")},
			{token.Whitespace, " ", pos("9:3")},
			{token.Implements, "implements", pos("9:4")},
			{token.Whitespace, " ", pos("9:14")},
			{token.Instanceof, "instanceof", pos("9:15")},
			{token.Whitespace, " ", pos("9:25")},
			{token.Insteadof, "insteadof", pos("9:26")},
			{token.Whitespace, " ", pos("9:35")},
			{token.Interface, "interface", pos("9:36")},
			{token.Whitespace, "\n", pos("9:45")},
			{token.Namespace, "namespace", pos("10:1")},
			{token.Whitespace, " ", pos("10:10")},
			{token.New, "new", pos("10:11")},
			{token.Whitespace, "\n", pos("10:14")},
			{token.Parent, "parent", pos("11:1")},
			{token.Whitespace, " ", pos("11:7")},
			{token.Private, "private", pos("11:8")},
			{token.Whitespace, " ", pos("11:15")},
			{token.Protected, "protected", pos("11:16")},
			{token.Whitespace, " ", pos("11:25")},
			{token.Public, "public", pos("11:26")},
			{token.Whitespace, "\n", pos("11:32")},
			{token.Return, "return", pos("12:1")},
			{token.Whitespace, "\n", pos("12:7")},
			{token.Self, "self", pos("13:1")},
			{token.Whitespace, " ", pos("13:5")},
			{token.Static, "static", pos("13:6")},
			{token.Whitespace, " ", pos("13:12")},
			{token.Switch, "switch", pos("13:13")},
			{token.Whitespace, "\n", pos("13:19")},
			{token.Throw, "throw", pos("14:1")},
			{token.Whitespace, " ", pos("14:6")},
			{token.Trait, "trait", pos("14:7")},
			{token.Whitespace, " ", pos("14:12")},
			{token.Try, "try", pos("14:13")},
			{token.Whitespace, "\n", pos("14:16")},
			{token.Use, "use", pos("15:1")},
			{token.Whitespace, "\n", pos("15:4")},
			{token.While, "while", pos("16:1")},
			{token.Whitespace, "\n", pos("16:6")},
			{token.EOF, "", pos("17:1")},
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
