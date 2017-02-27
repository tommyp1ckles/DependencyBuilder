# DependencyBuilder
Builds dependencies in topological order.

## Usage:
```Go
builder := dg.New()

// Add dependency type and it's constructor.
dg.AddDep(reflect.TypeOf(dep0{}), dep0Constructor)

// Build dependencies
dg.Build()
```
