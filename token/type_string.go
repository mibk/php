// Code generated by "stringer -type Type -linecomment"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Illegal-0]
	_ = x[EOF-1]
	_ = x[Whitespace-2]
	_ = x[Comment-3]
	_ = x[DocComment-4]
	_ = x[Ident-5]
	_ = x[String-6]
	_ = x[Var-7]
	_ = x[InlineHTML-8]
	_ = x[symbolStart-9]
	_ = x[OpenTag-10]
	_ = x[CloseTag-11]
	_ = x[Dollar-12]
	_ = x[Backslash-13]
	_ = x[Qmark-14]
	_ = x[Lparen-15]
	_ = x[Rparen-16]
	_ = x[Lbrack-17]
	_ = x[Rbrack-18]
	_ = x[Lbrace-19]
	_ = x[Rbrace-20]
	_ = x[Assign-21]
	_ = x[Lt-22]
	_ = x[Gt-23]
	_ = x[Period-24]
	_ = x[Comma-25]
	_ = x[Colon-26]
	_ = x[Semicolon-27]
	_ = x[Ellipsis-28]
	_ = x[Or-29]
	_ = x[And-30]
	_ = x[Quo-31]
	_ = x[Shl-32]
	_ = x[Shr-33]
	_ = x[symbolEnd-34]
	_ = x[keywordStart-35]
	_ = x[Abstract-36]
	_ = x[As-37]
	_ = x[Break-38]
	_ = x[Callable-39]
	_ = x[Case-40]
	_ = x[Catch-41]
	_ = x[Class-42]
	_ = x[Clone-43]
	_ = x[Const-44]
	_ = x[Continue-45]
	_ = x[Default-46]
	_ = x[Do-47]
	_ = x[Else-48]
	_ = x[Elseif-49]
	_ = x[Extends-50]
	_ = x[Final-51]
	_ = x[Finally-52]
	_ = x[Fn-53]
	_ = x[For-54]
	_ = x[Foreach-55]
	_ = x[Function-56]
	_ = x[Goto-57]
	_ = x[If-58]
	_ = x[Implements-59]
	_ = x[Instanceof-60]
	_ = x[Insteadof-61]
	_ = x[Interface-62]
	_ = x[Namespace-63]
	_ = x[New-64]
	_ = x[Parent-65]
	_ = x[Private-66]
	_ = x[Protected-67]
	_ = x[Public-68]
	_ = x[Return-69]
	_ = x[Self-70]
	_ = x[Static-71]
	_ = x[Switch-72]
	_ = x[Throw-73]
	_ = x[Trait-74]
	_ = x[Try-75]
	_ = x[Use-76]
	_ = x[While-77]
	_ = x[keywordEnd-78]
}

const _Type_name = "IllegalEOFWhitespaceCommentDocCommentIdentStringVarInlineHTMLsymbolStart<?php?>$\\?()[]{}=<>.,:;...|&/<<>>symbolEndkeywordStartabstractasbreakcallablecasecatchclasscloneconstcontinuedefaultdoelseelseifextendsfinalfinallyfnforforeachfunctiongotoifimplementsinstanceofinsteadofinterfacenamespacenewparentprivateprotectedpublicreturnselfstaticswitchthrowtraittryusewhilekeywordEnd"

var _Type_index = [...]uint16{0, 7, 10, 20, 27, 37, 42, 48, 51, 61, 72, 77, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 98, 99, 100, 101, 103, 105, 114, 126, 134, 136, 141, 149, 153, 158, 163, 168, 173, 181, 188, 190, 194, 200, 207, 212, 219, 221, 224, 231, 239, 243, 245, 255, 265, 274, 283, 292, 295, 301, 308, 317, 323, 329, 333, 339, 345, 350, 355, 358, 361, 366, 376}

func (i Type) String() string {
	if i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
