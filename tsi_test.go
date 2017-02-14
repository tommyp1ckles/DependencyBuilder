package dg

import (
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"os"
	"reflect"
	"testing"
)

var (
	logger = log.New(os.Stderr, "LOG: ", 0)
)
func TestDependencyBuilder(t *testing.T) {
	Convey("Testing Dependency Builder", t, func() {
		// 'Good' graphs
		Convey("Graph 1", func() {
			// Test 1
			//   .-> 1-----+
			//  /          v
			// 0           3
			//  \          ^
			//   .-> 2 ----+
			type static0 struct{}

			type dep0 struct{}

			cdep0 := func(s static0) dep0 {
				logger.Println("constructing dep0", s)
				return dep0{}
			}

			type dep1 struct{}

			cdep1 := func(r0 dep0) dep1 {
				logger.Println("constructing dep1", r0)
				return dep1{}
			}

			type dep2 struct{}

			cdep2 := func(r0 dep0) dep2 {
				logger.Println("constructing dep2", r0)
				return dep2{}
			}

			type dep3 struct{}

			cdep3 := func(r0 dep1, r1 dep2) dep3 {
				logger.Println("constructing dep3", r0, r1)
				return dep3{}
			}
			Convey("Test incomplete dependency graphs 1 (static)", func() {
				dg := New()
				dg.AddDep(reflect.TypeOf(dep0{}), cdep0)
				dg.AddDep(reflect.TypeOf(dep1{}), cdep1)
				dg.AddDep(reflect.TypeOf(dep2{}), cdep2)
				dg.AddDep(reflect.TypeOf(dep3{}), cdep3)

				Convey("Should fail due to missing dependency", func() {
					// Should fail due to missing static dep 0.
					So(dg.Build(), ShouldEqual, ErrIncompleteDependencyGraph)
				})
			})

			Convey("Test incomplete dependency graphs 1 (dep)", func() {
				dg := New()
				dg.AddStatic(static0{})
				dg.AddDep(reflect.TypeOf(dep0{}), cdep0)
				dg.AddDep(reflect.TypeOf(dep1{}), cdep1)
				//dg.AddDep(reflect.TypeOf(dep2{}), cdep2)
				dg.AddDep(reflect.TypeOf(dep3{}), cdep3)

				Convey("Should fail due to missing dependency", func() {
					// Should fail due to missing dep 2
					So(dg.Build(), ShouldEqual, ErrIncompleteDependencyGraph)
				})
			})

			Convey("Test complete dependency graph 1", func() {
				dg := New()
				dg.AddStatic(static0{})
				dg.AddDep(reflect.TypeOf(dep0{}), cdep0)
				dg.AddDep(reflect.TypeOf(dep1{}), cdep1)
				dg.AddDep(reflect.TypeOf(dep2{}), cdep2)
				dg.AddDep(reflect.TypeOf(dep3{}), cdep3)

				Convey("Should succeed", func() {
					So(dg.Build(), ShouldBeNil)
				})

			})
		})
		// Graph 2
		Convey("Graph 2", func() {
			// Test 2
			// static0 --> 3 --> 2 --> 1 --> 0
			// static1 ----------------------^
			type static0 struct{}
			type dep3 struct{}
			cdep3 := func(s static0) dep3 {
				return dep3{}
			}
			type dep2 struct{}
			cdep2 := func(r0 dep3) dep2 {
				return dep2{}
			}
			type dep1 struct{}
			cdep1 := func(r0 dep2) dep1 {
				return dep1{}
			}
			type dep0 struct{}
			cdep0 := func(r0 dep1, s static0) dep0 {
				return dep0{}
			}

			Convey("Should succeed", func() {
				dg := New()
				dg.AddStatic(static0{})
				dg.AddDep(reflect.TypeOf(dep1{}), cdep1)
				dg.AddDep(reflect.TypeOf(dep0{}), cdep0)
				dg.AddDep(reflect.TypeOf(dep3{}), cdep3)
				dg.AddDep(reflect.TypeOf(dep2{}), cdep2)
				So(dg.Build(), ShouldBeNil)
			})
		})
		// Cycle 1
		Convey("Cycle 1", func() {
			// 0 <-- 1
			// |     ^
			// V     |
			// 3 --> 2
			type dep0 struct{}
			type dep3 struct{}
			cdep3 := func(r0 dep0) dep3 {
				return dep3{}
			}
			type dep2 struct{}
			cdep2 := func(r0 dep3) dep2 {
				return dep2{}
			}
			type dep1 struct{}
			cdep1 := func(r0 dep2) dep1 {
				return dep1{}
			}
			cdep0 := func(r0 dep1) dep0 {
				return dep0{}
			}

			Convey("Should fail due to cycle", func() {
			    dg := New()
			    dg.AddDep(reflect.TypeOf(dep1{}), cdep1)
			    dg.AddDep(reflect.TypeOf(dep0{}), cdep0)
			    dg.AddDep(reflect.TypeOf(dep3{}), cdep3)
			    dg.AddDep(reflect.TypeOf(dep2{}), cdep2)
			    So(dg.Build(), ShouldEqual, ErrNotDAG)
			})
		})
	})
}
