# myth

A dependency-free, data-oriented sparse-set Entity Component System (ECS) for
Go 1.22+. Components live in contiguous typed slices for cache-friendly iteration.
The storage is lock-free in the informal, mutex-free sense: it contains no locks,
but is **not a concurrent lock-free algorithm**. Use one owning goroutine or
externally serialize all access that could race with mutation.

## Architecture

```text
                    Parent application
                  game loop and systems
                            |
                            v
                         World
                central lifecycle hub
                 /                  \
                v                    v
         EntityManager       registered UntypedPool list
         - allocate IDs               |
         - generations          +-----+-----------+
         - recycle IDs          |                 |
                                v                 v
                         Pool[Position]    Pool[Velocity]
                         dense components  dense components
                         dense entity IDs  dense entity IDs
                         generations       generations
                         sparse indices    sparse indices
                                \                 /
                                 v               v
                              Query2 ... Query5
                           smallest-pool iteration
```

`World.Create` delegates to its entity manager. `World.Destroy` removes the
entity's matching components from every registered pool before recycling its ID.
Pools remain typed in the application; `UntypedPool` exposes only `Remove` and
`Clear` for lifecycle coordination. Register each pool once per world.

## Consuming Parent Application

For a local checkout, put your application next to the `myth` directory and run
these commands from the application directory:

```sh
go mod init example.com/game
go mod edit -require=myth@v0.0.0 -replace=myth=../myth
```

The replacement path must point to the checkout containing this package's
module definition. The following complete program can be used as the parent
application's main package. Run it with `go run .`.

```go
package main

import (
	"fmt"

	"myth"
)

type Position struct {
	X, Y float64
}

type Velocity struct {
	X, Y float64
}

func main() {
	world := myth.NewWorld(1024)
	positions := myth.NewPool[Position](1024)
	velocities := myth.NewPool[Velocity](1024)
	world.RegisterPool(positions)
	world.RegisterPool(velocities)

	entity := world.Create()
	positions.Add(entity, Position{X: 10, Y: 20})
	velocities.Add(entity, Velocity{X: 2, Y: 4})

	stationary := world.Create()
	positions.Add(stationary, Position{X: 100, Y: 200})

	const deltaTime = 1.0 / 60.0
	for tick := 0; tick < 3; tick++ {
		myth.Query2(positions, velocities, func(entity myth.Entity, position *Position, velocity *Velocity) {
			position.X += velocity.X * deltaTime
			position.Y += velocity.Y * deltaTime
		})
	}

	if position := positions.Get(entity); position != nil {
		fmt.Printf("Position: %.1f, %.1f\n", position.X, position.Y)
	}

	world.Destroy(entity)
	world.Destroy(stationary)
	fmt.Println("Components removed:", !positions.Has(entity) && !velocities.Has(entity))
}
```

Output:

```text
Position: 10.1, 20.2
Components removed: true
```

Only the moving entity matches `Query2`; the position-only entity is untouched.
No manual pool removal is needed when destroying an entity through its world.

## Pool Operations

| Operation | Behavior |
| --- | --- |
| `NewPool[T](capacity)` | Reserves dense and sparse storage; nonpositive capacity selects a default. |
| `Add(entity, component)` | Inserts or replaces the component for the ID, recording the supplied generation. |
| `Get(entity)` | Returns a component pointer, or nil for a missing ID or stale generation. |
| `Has(entity)` | Performs the same membership checks as `Get`, returning a boolean. |
| `Remove(entity)` | Swap-removes a matching component and repairs the sparse index. |
| `Clear()` | Resets dense lengths and marks every sparse slot empty, preserving backing arrays. |

Sparse entries are `int32` dense indices, with `-1` marking empty slots.
Lookup, replacement, and swap removal are O(1). Insertion is amortized O(1) for
sequential IDs; adding a distant ID initializes all intervening sparse slots.
Sparse memory is proportional to the largest ID added, not just component count.
Dense indices must fit in a nonnegative `int32`; insertion panics on exhaustion.

`Pool.Clear` is O(sparse length), does not allocate replacement arrays, and does
not zero component values in retained storage. Pointer-bearing components can
therefore keep referenced objects alive until their slots are overwritten or the
pool is released. It does not force a garbage collection.

`World.Clear` calls `Clear` on every registered pool. It preserves registrations,
entity liveness, generations, and the manager's ID state. It is a component reset,
not an entity reset; existing entities can receive components again afterward.

## Ownership And Lifetime

- Use one world per simulation and keep its entities and pools together. Handles
  do not carry world identity, so cross-world IDs and generations can collide.
- Pass only live entities to `Add`. Pools cannot independently verify manager
  liveness. Register pools before relying on automatic component cleanup.
- Unregistered pools are not affected by `World.Destroy` or `World.Clear`.
  Standalone `EntityManager.Destroy` also does not clean up pools.
- Never copy initialized worlds, managers, or pools. Their zero values are usable.
- Component pointers are temporary. Do not retain them across `Add`, `Remove`,
  or `Clear`, and do not depend on dense ordering after swap removal.
- Nil pointer receivers are harmless. A nil world returns the zero entity from
  `Create`; a nil pool returns nil from `Get` and false from `Has`.
- Nil pool interfaces are ignored by registration. Typed nil `*Pool[T]` values
  are safe because pool methods handle nil receivers; custom `UntypedPool`
  implementations must provide their own receiver safety.

Entities start at generation 1; the zero entity is invalid for an entity manager.
Recycled IDs increment their generation. Stale or repeated destruction cannot
remove a newer generation's components or recycle its live ID. IDs at the maximum
16-bit generation are permanently retired to prevent generation wraparound.
Entity allocation panics if the 32-bit ID space is exhausted or allocation fails.

## Queries

`Query2`, `Query3`, `Query4`, and `Query5` iterate the smallest participating pool
sequentially and use generation-aware `Get` lookups on the remaining pools.
Only entities present with the same generation in every pool are yielded.
Ties favor the earliest pool argument. Queries allocate no intermediate storage.

Callbacks receive `Entity` followed by typed component pointers in pool argument
order. Values may be updated in place. Do not add, remove, or clear participating
pools, destroy entities through the world, or otherwise structurally mutate those
pools during a query. Defer those changes until iteration finishes. Nil pools or
callbacks produce no callbacks, as do empty participating pools.

## Validation

```sh
go test ./...
go vet ./...
```