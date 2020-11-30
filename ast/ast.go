package ast

import "mibk.io/php/token"

type File struct {
	Namespace *Name
	UseStmts  []*UseStmt
	Decls     []Decl
}

type UseStmt struct {
	Name *Name
}

type TokenBlob struct {
	Toks []token.Token
	Body *BlockStmt
}

type Decl interface{}

type ConstDecl struct {
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
	Body   *BlockStmt
}

type Param struct {
	Name string
}

type Stmt interface{}

type BlockStmt struct {
	List []Stmt
}

type Expr interface{}

// A Name represents a (possibly qualified or fully qualified) PHP
// name, which might be a class name, a built-in type, or a special
// value type (e.g. null, false).
type Name struct {
	Parts  []string
	Global bool // fully qualified
}
