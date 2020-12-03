package ast_test

import (
	"fmt"
	"strings"
	"testing"

	"mibk.io/php/ast"
)

func TestSyntaxErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{{
		"unterminated comment",
		"<?php /**",
		`syntax:1:10: unterminated block comment`,
	}, {
		"unterminated string",
		"<?php '",
		`syntax:1:8: string not terminated`,
	}, {
		"wrong open tag",
		`   <?php `,
		`syntax:1:1: expecting <?php, found InlineHTML("   ")`,
	}, {
		"invalid PHPDoc",
		"<?php\n   /** @var */",
		`syntax:2:13: parsing PHPDoc: expecting ( or basic type, found */`,
	}, {
		"unexpected /",
		`<?php /`,
		`syntax:1:7: unexpected /`,
	}, {
		"unterminated param list",
		`<?php function a(`,
		`syntax:1:18: expecting ), found EOF`,
	}, {
		"unterminated class",
		`<?php class a{`,
		`syntax:1:15: expecting }, found EOF`,
	}, {
		"missing default",
		`<?php function a($x=,`,
		`syntax:1:21: unexpected ,, expecting lit`,
	}}

	for _, tt := range tests {
		file, err := ast.Parse(strings.NewReader(tt.input))
		errStr := "<nil>"
		if err != nil {
			if file != nil {
				t.Fatalf("%q: got %+v on err", tt.input, file)
			}
			if se, ok := err.(*ast.SyntaxError); ok {
				err = fmt.Errorf("syntax:%d:%d: %v", se.Line, se.Column, se.Err)
			}
			errStr = err.Error()
		}
		if errStr != tt.wantErr {
			t.Errorf("%s:\n got %s\nwant %s", tt.name, errStr, tt.wantErr)
		}
	}
}
