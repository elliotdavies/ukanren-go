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

	res1, ok := unify(Var{1}, Lit(3), map1)
	fmt.Print("Should have unified successfully: ")
	if ok {
		fmt.Println("correct", res1)
	} else {
		fmt.Println("Uh oh")
	}

	res2, ok := unify(Var{1}, Lit(4), map1)
	fmt.Print("Should not have unified successfully: ")
	if ok {
		fmt.Println("Uh oh")
	} else {
		fmt.Println("correct", res2)
	}

	fmt.Println("Eq goal - should see 1:1 in map:", *callGoal(eq(Var{1}, Lit(1))))

	lam := func(t Type) Goal {
		return eq(t, Lit(23))
	}
	res := callGoal(callFresh(lam))
	fmt.Println("callFresh eq goal - should see 1:1 in map:", *res)

	lam = func(t Type) Goal {
		return either(eq(t, Lit(23)), eq(t, Lit(42)))
	}
	disj := callGoal(callFresh(lam))
	fmt.Println("disj", head(*disj), *tail(*disj))

	lam = func(t Type) Goal {
		return both(eq(t, Lit(23)), eq(t, Lit(42)))
	}
	conj1 := callGoal(callFresh(lam))
	fmt.Println("conj1", conj1)

	lam = func(t Type) Goal {
		return both(eq(t, Lit(23)), eq(t, Lit(23)))
	}
	conj2 := callGoal(callFresh(lam))
	fmt.Println("conj2", *conj2)
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
	head Type
	tail *Cons
}

func (c Cons) isType() {}

func cons(head Type, tail *Cons) Cons {
	return Cons{head, tail}
}

func head(c Cons) Type {
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

// Try to unify two terms and update the substitution map accordingly
// In case of failure return nil
func unify(t1 Type, t2 Type, substMap Map) (Map, bool) {
	t1 = resolve(t1, substMap)
	t2 = resolve(t2, substMap)
	// If both equal, don't need to do anything
	if isVar(t1) && isVar(t2) && t1 == t2 {
		return substMap, true
	} else if isVar(t1) {
		// Extend map with mapping from t1 to t2
		return extend(substMap, t1, t2), true
	} else if isVar(t2) {
		// Extend map with mapping from t2 to t1
		return extend(substMap, t2, t1), true
	} else if isCons(t1) && isCons(t2) {
		// If both cons cells then try to unify elements in the lists
		// @TODO Scrap isCons method?
		t1, t1ok := t1.(Cons)
		t2, t2ok := t2.(Cons)
		if t1ok && t2ok {
			substMap, ok := unify(head(t1), head(t2), substMap)
			if ok {
				return unify(tail(t1), tail(t2), substMap)
			}
			return nil, false
		}
	} else if t1 == t2 {
		// If equal at this stage, no need to do anything
		return substMap, true
	}

	// Fail - could not unify
	return nil, false
}

// State
type State struct {
	substMap Map
	count    int
}

func (s State) isType() {}

// Streams
type Stream *Cons

func newStream(s State) Stream {
	stream := cons(s, nil)
	return &stream
}

// Goals
type Goal func(State) Stream

func callGoal(g Goal) *Cons {
	emptyState := State{make(Map), 0}
	return g(emptyState)
}

func callFresh(f func(Type) Goal) Goal {
	return func(s State) Stream {
		goal := f(Var{s.count})
		newState := State{s.substMap, s.count + 1}
		return goal(newState)
	}
}

func append(s1 Stream, s2 Stream) Stream {
	if s1 == nil {
		return s2
	}
	stream := cons(head(*s1), append(tail(*s1), s2))
	return &stream
}

func mappend(g Goal, s Stream) Stream {
	if s == nil {
		return nil
	}

	h := head(*s)
	state, ok := h.(State)
	if !ok {
		return nil
	}

	stream := append(g(state), mappend(g, tail(*s)))
	return stream
}

// Equality goal - if terms unify they are equal
func eq(t1 Type, t2 Type) Goal {
	return func(s State) Stream {
		substMap, ok := unify(t1, t2, s.substMap)
		if ok {
			return newStream(State{substMap, s.count})
		}
		return nil
	}
}

// Disjunction
func either(g1 Goal, g2 Goal) Goal {
	return func(s State) Stream {
		return append(g1(s), g2(s))
	}
}

// Conjunction
func both(g1 Goal, g2 Goal) Goal {
	return func(s State) Stream {
		return mappend(g2, g1(s))
	}
}
