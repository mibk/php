package ast

import (
	"mibk.io/php/token"
	"mibk.io/phpdoc"
)

type File struct {
	Namespace *Name
	UseStmts  []*UseStmt
	Decls     []Decl
}

type UseStmt struct {
	Name *Name
}

type Decl interface{ doc() *phpdoc.Block }

type ConstDecl struct {
	Doc  *phpdoc.Block // or nil
	Name string
	X    Expr
}

type ClassDecl struct {
	Doc     *phpdoc.Block // or nil
	Name    string
	Members []ClassMember
}

func (d *ConstDecl) doc() *phpdoc.Block { return d.Doc }
func (d *ClassDecl) doc() *phpdoc.Block { return d.Doc }

type ClassMember interface{ doc() *phpdoc.Block }

type MethodDecl struct {
	Doc    *phpdoc.Block // or nil
	Name   string
	Params []*Param
	Body   *BlockStmt
}

func (d *MethodDecl) doc() *phpdoc.Block { return d.Doc }

type Param struct {
	Name string
}

type Stmt interface{}

type UnknownStmt struct {
	Toks []token.Token
	Body *BlockStmt
}

type BlockStmt struct {
	List []Stmt
}

type Expr interface{}

type UnknownExpr struct {
	Toks []token.Token
}

// A Name represents a (possibly qualified or fully qualified) PHP
// name, which might be a class name, a built-in type, or a special
// value type (e.g. null, false).
type Name struct {
	Parts  []string
	Global bool // fully qualified
}
