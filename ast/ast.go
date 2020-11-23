package ast

import "mibk.io/php/token"

type File struct {
	Namespace *Name
	Decls     []Decl
}

type TokenBlob struct {
	Toks []token.Token
}

type Decl interface{}

type Const struct {
	Name string
	X    Expr
}

type ClassDecl struct {
	Name    string
	Members []ClassMember
}

type ClassMember interface{}

type Method struct {
	Name   string
	Params []*Param
}

type Param struct {
	Name string
}

type Expr interface{}

// A Name represents a (possibly qualified or fully qualified) PHP
// name, which might be a class name, a built-in type, or a special
// value type (e.g. null, false).
type Name struct {
	Parts  []string
	Global bool // fully qualified
}
