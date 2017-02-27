package dg

import (
	"errors"
	"reflect"
)

var (
	// ErrIncompleteDependencyGraph is returned when a dependency graph
	// is missing dependencies required to build dependencies.
	ErrIncompleteDependencyGraph = errors.New("Dependency graph is missing a prerequisite dependency")

	// ErrNotDAG is returned when trying to build a dependency graph
	// that contains a cycle (i.e. has circular dependencies).
	ErrNotDAG = errors.New("Circular dependecies")

	// ErrUnexpectedConstructorContext is returned when a constructor
	// function returns more than the expected values.
	// TODO: find a way to compensate for different return numbers.
	ErrUnexpectedConstructorContext = errors.New("Constructor returned an unexpected number of values")
)

// Dependency contains a needed interface its corresponding constructor.
type Dependency struct {
	Interface   reflect.Value
	Constructor interface{}
	needs       int /* how many non static deps does this need */
	neededBy    []*Dependency

	id string
}

// Build builds a dependency graph.
func (dg *DependencyGraph) Build() error {
	start, err := buildGraph(dg.deps, dg.static)
	if err != nil {
		return err
	}

	order, err := traverse(dg.deps, start)
	if err != nil {
		return err
	}

	return build(order, dg.deps, dg.static)
}

// constructs a dependency graph of the dependency map.
func buildGraph(
	deps map[reflect.Type]*Dependency, /* deps to be build */
	staticDeps map[reflect.Type]reflect.Value, /* deps already satsified */
) ([]*Dependency, error) {
	startNodes := []*Dependency{}
	for _, dep := range deps {
		cType := reflect.TypeOf(dep.Constructor)
		cParams := cType.NumIn() // Panics if c is not a function.
		for i := 0; i < cParams; i++ {
			aType := cType.In(i) // type of argument

			// These deps do not need to be constructed
			_, isStaticDep := staticDeps[aType]
			if isStaticDep {
				continue
			}

			// These do
			m, depExists := deps[aType]
			if !depExists {
				return nil, ErrIncompleteDependencyGraph
			}
			/* makes the prereq node point at this */
			m.neededBy = append(m.neededBy, dep)
			/* increments how many non static deps this node has.*/
			dep.needs++
		}
		if dep.needs == 0 { //not including static-deps.
			startNodes = append(startNodes, dep)
		}
	}
	return startNodes, nil
}

func hasEdges(deps map[reflect.Type]*Dependency) bool {
	for _, dep := range deps {
		if dep.needs != 0 {
			return true
		}
	}
	return false
}

// returns a topological ordering of a dependency graph.
func traverse(deps map[reflect.Type]*Dependency, startNodes []*Dependency) (
	[]*Dependency, error) {
	L := []*Dependency{}
	S := startNodes
	for {
		if len(S) == 0 {
			break
		}
		n := S[0]
		S = S[1:len(S)]
		L = append(L, n)
		for _, m := range n.neededBy {
			m.needs--
			if m.needs == 0 { // becomes a start node.
				S = append(S, m)
			}
		}
		n.neededBy = nil
	}
	if hasEdges(deps) {
		return nil, ErrNotDAG
	}
	return L, nil
}

// invokes the constructors in the order specified and populates the
// dependencies.
func build(
	L []*Dependency,
	deps map[reflect.Type]*Dependency,
	staticDeps map[reflect.Type]reflect.Value,
) error {
	for _, dep := range L {
		cVal := reflect.ValueOf(dep.Constructor)
		cType := reflect.TypeOf(dep.Constructor)
		args := make([]reflect.Value, cType.NumIn())

		for i := 0; i < cType.NumIn(); i++ {
			aType := cType.In(i) // type of argument
			dv, ok := staticDeps[aType]
			if ok {
				args[i] = dv
				continue
			}

			d, ok := deps[aType]
			if ok {
				args[i] = d.Interface
			}
		}

		rets := cVal.Call(args)
		if len(rets) != 1 {
			// TODO: Figure out a way to handle this.
			return ErrUnexpectedConstructorContext
		}
		dep.Interface = rets[0]
	}
	return nil
}

// DependencyGraph
type DependencyGraph struct {
	deps   map[reflect.Type]*Dependency
	static map[reflect.Type]reflect.Value
}

// New creates a new dependency graph.
func New() *DependencyGraph {
	return &DependencyGraph{
		deps:   make(map[reflect.Type]*Dependency),
		static: make(map[reflect.Type]reflect.Value),
	}
}

// AddDep adds a new dependency.
func (dg *DependencyGraph) AddDep(i reflect.Type, c interface{}) {
	dg.deps[i] = &Dependency{
		Constructor: c,
		needs:       0,
		neededBy:    []*Dependency{},
	}
}

// AddStatic adds a new static dependency.
func (dg *DependencyGraph) AddStatic(dep interface{}) {
	dg.static[reflect.TypeOf(dep)] = reflect.ValueOf(dep)
}
