package ast

import (
	"mibk.io/phpdoc"
)

type File struct {
	Pragmas   []*Pragma
	Namespace *Name
	UseStmts  []*UseStmt
	Stmts     []Stmt
}

// TODO: Pragma.X Expr?

type Pragma struct {
	Name  string
	Value Expr
}

type UseStmt struct {
	Name  *Name
	Alias string // or ""
}

type Decl interface{ doc() *phpdoc.Block }

type ConstDecl struct {
	Doc  *phpdoc.Block // or nil
	Name string
	X    Expr
}

type VarDecl struct {
	Doc    *phpdoc.Block // or nil
	Name   string
	Static bool // valid for class props
	X      Expr
}

type FuncDecl struct {
	Doc    *phpdoc.Block // or nil
	Name   string
	Static bool // valid for methods
	Params []*Param
	Result *Type      // or nil
	Body   *BlockStmt // or nil (e.g. interfaces)
}

type Param struct {
	Type     *Type // or nil
	ByRef    bool  // pass by reference
	Variadic bool
	Name     string
	Default  Expr // or nil
}

// TODO: Make abstract and final mutually exclusive?

type ClassDecl struct {
	Doc        *phpdoc.Block // or nil
	Name       string
	Abstract   bool
	Final      bool
	Extends    *Name // or nil
	Implements []*Name
	Traits     []*UseStmt
	Members    []*ClassMember
}

type InterfaceDecl struct {
	Doc     *phpdoc.Block // or nil
	Name    string
	Extends *Name // or nil
	Members []*ClassMember
}

type TraitDecl struct {
	Doc     *phpdoc.Block // or nil
	Name    string
	Members []*ClassMember
}

func (d *ConstDecl) doc() *phpdoc.Block     { return d.Doc }
func (d *VarDecl) doc() *phpdoc.Block       { return d.Doc }
func (d *FuncDecl) doc() *phpdoc.Block      { return d.Doc }
func (d *ClassDecl) doc() *phpdoc.Block     { return d.Doc }
func (d *InterfaceDecl) doc() *phpdoc.Block { return d.Doc }
func (d *TraitDecl) doc() *phpdoc.Block     { return d.Doc }

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
	X    Expr
	Body *BlockStmt
}

type BlockStmt struct {
	List []Stmt
}

// TODO: Init and Post should be statements.

type ForStmt struct {
	Init Expr // or nil
	Cond Expr // or nil
	Post Expr // or nil
	Body Stmt
}

type Expr interface{}

type StaticSelectorExpr struct {
	X   Expr
	Sel string
}

type ArrayLit struct {
	Elems []Expr
}

// TODO: Separate type for scope?

type FuncLit struct {
	Params []*Param
	Scope  []*Param
	Result *Type // or nil
	Body   *BlockStmt
}

type UnknownExpr struct {
	Elems []interface{}
}

type Type struct {
	Nullable bool
	Name     *Name
}

// A Name represents a (possibly qualified or fully qualified) PHP
// name, which might be a class name, a built-in type, or a special
// value type (e.g. null, false).
type Name struct {
	Parts  []string
	Global bool // fully qualified
}
