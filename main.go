package main

import "fmt"

func main() {
	lit1 := Lit(1)
	var1 := Var{1}
	cons1 := Cons{Lit(2), nil}

	fmt.Println("lit1 is a var:", isVar(lit1))
	fmt.Println("var1 is a var:", isVar(var1))
	fmt.Println("cons1:", cons1)
	fmt.Println("cons1 head:", head(cons1))
	fmt.Println("cons1 tail:", tail(cons1))
	fmt.Println("var1 is a cons:", isCons(var1))
	fmt.Println("cons1 is a cons:", isCons(cons1))

	map1 := make(Map)
	map1[Var{1}] = Var{2}
	map1[Var{2}] = Lit(3)
	fmt.Println("Should be 3:", resolve(Var{1}, map1))
}

// Nearest thing to a sum type we're going to get
type Type interface {
	isType()
}

// Wrap literals because otherwise can't make them instances of Type
type Lit int

func (l Lit) isType() {}

// Vars
type Var struct {
	v int
}

func (v Var) isType() {}

func isVar(v Type) bool {
	switch v.(type) {
	case Var:
		return true
	default:
		return false
	}
}

// Cons cells
type Cons struct {
	head Lit
	tail *Cons
}

func (c Cons) isType() {}

func cons(head Lit, tail *Cons) Cons {
	return Cons{head, tail}
}

func head(c Cons) Lit {
	return c.head
}

func tail(c Cons) *Cons {
	return c.tail
}

func isCons(v Type) bool {
	switch v.(type) {
	case Cons:
		return true
	default:
		return false
	}
}

// Maps
type Map map[Type]Type

func extend(m Map, key Type, val Type) Map {
	copy := make(Map)
	for k, v := range m {
		copy[k] = v
	}
	copy[key] = val
	return copy
}

// Work out what a term equates to
func resolve(term Type, substMap Map) Type {
	if isVar(term) && substMap[term] != nil {
		return resolve(substMap[term], substMap)
	}
	return term
}
