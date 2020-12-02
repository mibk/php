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

type VarDecl struct {
	Doc  *phpdoc.Block // or nil
	Name string
	X    Expr
}

type FuncDecl struct {
	Doc    *phpdoc.Block // or nil
	Name   string
	Params []*Param
	Body   *BlockStmt
}

type Param struct {
	Name string
}

type ClassDecl struct {
	Doc     *phpdoc.Block // or nil
	Name    string
	Traits  []*UseStmt
	Members []*ClassMember
}

func (d *ConstDecl) doc() *phpdoc.Block { return d.Doc }
func (d *VarDecl) doc() *phpdoc.Block   { return d.Doc }
func (d *FuncDecl) doc() *phpdoc.Block  { return d.Doc }
func (d *ClassDecl) doc() *phpdoc.Block { return d.Doc }

type Vis uint

const (
	DefaultVis Vis = iota
	Public
	Protected
	Private
)

type ClassMember struct {
	Doc  *phpdoc.Block // or nil
	Vis  Vis
	Decl Decl
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
