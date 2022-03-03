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
actually have to be a <html>
<?phpnamespace <?php`,
		[]token.Token{
			{token.InlineHTML, "doesn't\nactually have to be a <html>\n<?phpnamespace <?php", pos("1:1")},
			{token.EOF, "", pos("3:21")},
		},
	}, {
		"tease opening",
		`< <?ph  <?p <?hp nic <?php `,
		[]token.Token{
			{token.InlineHTML, "< <?ph  <?p <?hp nic ", pos("1:1")},
			{token.OpenTag, "<?php", pos("1:22")},
			{token.Whitespace, " ", pos("1:27")},
			{token.EOF, "", pos("1:28")},
		},
	}, {
		"basic PHP",
		`<html> <?php

   echo 'ahoj';?>
<?php endif`,
		[]token.Token{
			{token.InlineHTML, "<html> ", pos("1:1")},
			{token.OpenTag, "<?php", pos("1:8")},
			{token.Whitespace, "\n\n   ", pos("1:13")},
			{token.Ident, "echo", pos("3:4")},
			{token.Whitespace, " ", pos("3:8")},
			{token.String, `'ahoj'`, pos("3:9")},
			{token.Semicolon, ";", pos("3:15")},
			{token.CloseTag, "?>", pos("3:16")},
			{token.InlineHTML, "\n", pos("3:18")},
			{token.OpenTag, "<?php", pos("4:1")},
			{token.Whitespace, " ", pos("4:6")},
			{token.Ident, "endif", pos("4:7")},
			{token.EOF, "", pos("4:12")},
		},
	}, {
		"comments",
		`<?php // line comment
namespace /*block ?> */ DateTime/** comments*/;# another line comm? or?
// early ?><?php # eof`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.Comment, "// line comment", pos("1:7")},
			{token.Whitespace, "\n", pos("1:22")},
			{token.Namespace, "namespace", pos("2:1")},
			{token.Whitespace, " ", pos("2:10")},
			{token.Comment, "/*block ?> */", pos("2:11")},
			{token.Whitespace, " ", pos("2:24")},
			{token.Ident, "DateTime", pos("2:25")},
			{token.DocComment, "/** comments*/", pos("2:33")},
			{token.Semicolon, ";", pos("2:47")},
			{token.Comment, "# another line comm? or?", pos("2:48")},
			{token.Whitespace, "\n", pos("2:72")},
			{token.Comment, "// early ", pos("3:1")},
			{token.CloseTag, "?>", pos("3:10")},
			{token.OpenTag, "<?php", pos("3:12")},
			{token.Whitespace, " ", pos("3:17")},
			{token.Comment, "# eof", pos("3:18")},
			{token.EOF, "", pos("3:23")},
		},
	}, {
		"misc",
		`<?php &....|..`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.And, "&", pos("1:7")},
			{token.Ellipsis, "...", pos("1:8")},
			{token.Period, ".", pos("1:11")},
			{token.Or, "|", pos("1:12")},
			{token.Illegal, "..", pos("1:13")},
			{token.EOF, "", pos("1:15")},
		},
	}, {
		"single quoted strings",
		`<?php '\'\\' '\\' '\'' '\\n\\\'''
\'\@'`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.String, `'\'\\'`, pos("1:7")},
			{token.Whitespace, " ", pos("1:13")},
			{token.String, `'\\'`, pos("1:14")},
			{token.Whitespace, " ", pos("1:18")},
			{token.String, `'\''`, pos("1:19")},
			{token.Whitespace, " ", pos("1:23")},
			{token.String, `'\\n\\\''`, pos("1:24")},
			{token.String, "'\n\\'\\@'", pos("1:33")},
			{token.EOF, "", pos("2:6")},
		},
	}, {
		"double quoted strings",
		`<?php "\"\\" "\\" "\"" "\\'\\\"""
\""
"\n\r\t\v\e\f\$\xED\u{2030}\%"`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.String, `"\"\\"`, pos("1:7")},
			{token.Whitespace, " ", pos("1:13")},
			{token.String, `"\\"`, pos("1:14")},
			{token.Whitespace, " ", pos("1:18")},
			{token.String, `"\""`, pos("1:19")},
			{token.Whitespace, " ", pos("1:23")},
			{token.String, `"\\'\\\""`, pos("1:24")},
			{token.String, "\"\n\\\"\"", pos("1:33")},
			{token.Whitespace, "\n", pos("2:4")},
			{token.String, "\"\\n\\r\\t\\v\\e\\f\\$\\xED\\u{2030}\\%\"", pos("3:1")},
			{token.EOF, "", pos("3:31")},
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
		`<?php <><<>>`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.Lt, "<", pos("1:7")},
			{token.Gt, ">", pos("1:8")},
			{token.Shl, "<<", pos("1:9")},
			{token.Shr, ">>", pos("1:11")},
			{token.EOF, "", pos("1:13")},
		},
	}, {
		"heredoc",
		`<?php <<<	 END ` + `
buffalo
  END
END:
END;nic
END;	` + `
<<<"HERE"
there
HERE

(<<<XX
x
XX,
)
`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.String, "<<<\t END \nbuffalo\n  END\nEND:\nEND;nic\nEND", pos("1:7")},
			{token.Semicolon, ";", pos("6:4")},
			{token.Whitespace, "\t\n", pos("6:5")},
			{token.String, "<<<\"HERE\"\nthere\nHERE", pos("7:1")},
			{token.Whitespace, "\n\n", pos("9:5")},
			{token.Lparen, "(", pos("11:1")},
			{token.String, "<<<XX\nx\nXX", pos("11:2")},
			{token.Comma, ",", pos("13:3")},
			{token.Whitespace, "\n", pos("13:4")},
			{token.Rparen, ")", pos("14:1")},
			{token.Whitespace, "\n", pos("14:2")},
			{token.EOF, "", pos("15:1")},
		},
	}, {
		"nowdoc",
		`<?php	<<<	 'NOWdoc' ` + `
weather
  NOWdoc
NOWdoc:
NOWdoc;nada
NOWdoc;	` + `
`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, "\t", pos("1:6")},
			{token.String, "<<<\t 'NOWdoc' \nweather\n  NOWdoc\nNOWdoc:\nNOWdoc;nada\nNOWdoc", pos("1:7")},
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
declare default do
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
			{token.Ident, "callable", pos("4:1")},
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
			{token.Declare, "declare", pos("5:1")},
			{token.Whitespace, " ", pos("5:8")},
			{token.Default, "default", pos("5:9")},
			{token.Whitespace, " ", pos("5:16")},
			{token.Do, "do", pos("5:17")},
			{token.Whitespace, "\n", pos("5:19")},
			{token.Else, "else", pos("6:1")},
			{token.Whitespace, " ", pos("6:5")},
			{token.Else, "else", pos("6:6")},
			{token.If, "if", pos("6:10")},
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
			{token.Ident, "parent", pos("11:1")},
			{token.Whitespace, " ", pos("11:7")},
			{token.Private, "private", pos("11:8")},
			{token.Whitespace, " ", pos("11:15")},
			{token.Protected, "protected", pos("11:16")},
			{token.Whitespace, " ", pos("11:25")},
			{token.Public, "public", pos("11:26")},
			{token.Whitespace, "\n", pos("11:32")},
			{token.Return, "return", pos("12:1")},
			{token.Whitespace, "\n", pos("12:7")},
			{token.Ident, "self", pos("13:1")},
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
	}, {
		"numbers",
		`<?php
0 07 007 34487908803190 0xff 0XFA 0b10
3.14 0.09 -0.0014e-13-.14
`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, "\n", pos("1:6")},
			{token.Int, "0", pos("2:1")},
			{token.Whitespace, " ", pos("2:2")},
			{token.Int, "07", pos("2:3")},
			{token.Whitespace, " ", pos("2:5")},
			{token.Int, "007", pos("2:6")},
			{token.Whitespace, " ", pos("2:9")},
			{token.Int, "34487908803190", pos("2:10")},
			{token.Whitespace, " ", pos("2:24")},
			{token.Int, "0xff", pos("2:25")},
			{token.Whitespace, " ", pos("2:29")},
			{token.Int, "0XFA", pos("2:30")},
			{token.Whitespace, " ", pos("2:34")},
			{token.Int, "0b10", pos("2:35")},
			{token.Whitespace, "\n", pos("2:39")},
			{token.Float, "3.14", pos("3:1")},
			{token.Whitespace, " ", pos("3:5")},
			{token.Float, "0.09", pos("3:6")},
			{token.Whitespace, " ", pos("3:10")},
			{token.Sub, "-", pos("3:11")},
			{token.Float, "0.0014e-13", pos("3:12")},
			{token.Sub, "-", pos("3:22")},
			{token.Float, ".14", pos("3:23")},
			{token.Whitespace, "\n", pos("3:26")},
			{token.EOF, "", pos("4:1")},
		},
	}, {
		"symbols",
		`<?php = > => - ->+-+:::,`,
		[]token.Token{
			{token.OpenTag, "<?php", pos("1:1")},
			{token.Whitespace, " ", pos("1:6")},
			{token.Assign, "=", pos("1:7")},
			{token.Whitespace, " ", pos("1:8")},
			{token.Gt, ">", pos("1:9")},
			{token.Whitespace, " ", pos("1:10")},
			{token.DoubleArrow, "=>", pos("1:11")},
			{token.Whitespace, " ", pos("1:13")},
			{token.Sub, "-", pos("1:14")},
			{token.Whitespace, " ", pos("1:15")},
			{token.Arrow, "->", pos("1:16")},
			{token.Add, "+", pos("1:18")},
			{token.Sub, "-", pos("1:19")},
			{token.Add, "+", pos("1:20")},
			{token.DoubleColon, "::", pos("1:21")},
			{token.Colon, ":", pos("1:23")},
			{token.Comma, ",", pos("1:24")},
			{token.EOF, "", pos("1:25")},
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
			if err := sc.Err(); err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("tokens don't match: (-got +want)\n%s", diff)
			}
		})
	}
}

func TestScanErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{{
		"unterminated block comment",
		`<?php /*
nic
 `,
		"line:3:2: unterminated block comment",
	}, {
		"unterminated single quoted",
		`<?php 'foooo…`,
		"line:1:14: string not terminated",
	}, {
		"invalid heredoc #1",
		`<?php <<<`,
		"line:1:10: missing opening heredoc identifier",
	}, {
		"invalid heredoc #2",
		`<?php <<< 8`,
		"line:1:11: invalid opening heredoc identifier",
	}, {
		"invalid heredoc #3",
		`<?php <<< HERE x`,
		"line:1:16: no newline after identifier in heredoc",
	}, {
		"invalid heredoc #4",
		`<?php <<< HERE
HERE ;
`,
		"line:3:1: heredoc not terminated",
	}, {
		"invalid heredoc #5",
		`<?php <<< "HERE
`,
		"line:2:1: quoted heredoc identifier not terminated",
	}, {
		"invalid heredoc #5",
		`<?php <<< 'HERE
`,
		"line:2:1: quoted heredoc identifier not terminated",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := token.NewScanner(strings.NewReader(tt.input))

			for sc.Next().Type != token.EOF {
			}
			errStr := "<nil>"
			if err := sc.Err(); err != nil {
				errStr = err.Error()
			}
			if errStr != tt.wantErr {
				t.Errorf("\n got %s\nwant %s", errStr, tt.wantErr)
			}
		})
	}
}

func TestBadReader(t *testing.T) {
	sc := token.NewScanner(new(badReader))
	for sc.Next().Type != token.EOF {
	}
	errStr := "<nil>"
	if err := sc.Err(); err != nil {
		errStr = err.Error()
	}
	const wantErr = "i'm fine"
	if errStr != wantErr {
		t.Errorf("\n got %s\nwant %s", errStr, wantErr)
	}
}

type badReader struct{}

func (badReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("i'm fine")
}
