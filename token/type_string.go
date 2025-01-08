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
	_ = x[Int-6]
	_ = x[Float-7]
	_ = x[String-8]
	_ = x[Var-9]
	_ = x[InlineHTML-10]
	_ = x[symbolStart-11]
	_ = x[OpenTag-12]
	_ = x[CloseTag-13]
	_ = x[Dollar-14]
	_ = x[Backslash-15]
	_ = x[Qmark-16]
	_ = x[Lparen-17]
	_ = x[Rparen-18]
	_ = x[Lbrack-19]
	_ = x[Rbrack-20]
	_ = x[Lbrace-21]
	_ = x[Rbrace-22]
	_ = x[Add-23]
	_ = x[Sub-24]
	_ = x[Assign-25]
	_ = x[Lt-26]
	_ = x[Gt-27]
	_ = x[Period-28]
	_ = x[Comma-29]
	_ = x[Colon-30]
	_ = x[DoubleColon-31]
	_ = x[Semicolon-32]
	_ = x[Ellipsis-33]
	_ = x[Or-34]
	_ = x[And-35]
	_ = x[Quo-36]
	_ = x[Shl-37]
	_ = x[Shr-38]
	_ = x[Arrow-39]
	_ = x[DoubleArrow-40]
	_ = x[symbolEnd-41]
	_ = x[keywordStart-42]
	_ = x[Abstract-43]
	_ = x[As-44]
	_ = x[Break-45]
	_ = x[Case-46]
	_ = x[Catch-47]
	_ = x[Class-48]
	_ = x[Clone-49]
	_ = x[Const-50]
	_ = x[Continue-51]
	_ = x[Declare-52]
	_ = x[Default-53]
	_ = x[Do-54]
	_ = x[Else-55]
	_ = x[Enum-56]
	_ = x[Extends-57]
	_ = x[Final-58]
	_ = x[Finally-59]
	_ = x[Fn-60]
	_ = x[For-61]
	_ = x[Foreach-62]
	_ = x[From-63]
	_ = x[Function-64]
	_ = x[Global-65]
	_ = x[Goto-66]
	_ = x[If-67]
	_ = x[Implements-68]
	_ = x[Instanceof-69]
	_ = x[Insteadof-70]
	_ = x[Interface-71]
	_ = x[Match-72]
	_ = x[Namespace-73]
	_ = x[New-74]
	_ = x[Private-75]
	_ = x[Protected-76]
	_ = x[Public-77]
	_ = x[Readonly-78]
	_ = x[Return-79]
	_ = x[Static-80]
	_ = x[Switch-81]
	_ = x[Throw-82]
	_ = x[Trait-83]
	_ = x[Try-84]
	_ = x[Use-85]
	_ = x[While-86]
	_ = x[Yield-87]
	_ = x[keywordEnd-88]
}

const _Type_name = "IllegalEOFWhitespaceCommentDocCommentIdentIntFloatStringVarInlineHTMLsymbolStart<?php?>$\\?()[]{}+-=<>.,:::;...|&/<<>>->=>symbolEndkeywordStartabstractasbreakcasecatchclasscloneconstcontinuedeclaredefaultdoelseenumextendsfinalfinallyfnforforeachfromfunctionglobalgotoifimplementsinstanceofinsteadofinterfacematchnamespacenewprivateprotectedpublicreadonlyreturnstaticswitchthrowtraittryusewhileyieldkeywordEnd"

var _Type_index = [...]uint16{0, 7, 10, 20, 27, 37, 42, 45, 50, 56, 59, 69, 80, 85, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 106, 107, 110, 111, 112, 113, 115, 117, 119, 121, 130, 142, 150, 152, 157, 161, 166, 171, 176, 181, 189, 196, 203, 205, 209, 213, 220, 225, 232, 234, 237, 244, 248, 256, 262, 266, 268, 278, 288, 297, 306, 311, 320, 323, 330, 339, 345, 353, 359, 365, 371, 376, 381, 384, 387, 392, 397, 407}

func (i Type) String() string {
	if i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
