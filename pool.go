package myth

// Pool stores components contiguously, with O(1) lookup and swap removal.
// The zero value is usable. A Pool must not be copied after first use.
// Mutations and access to returned pointers must be externally serialized.
type Pool[T any] struct {
	components  []T
	entities    []uint32
	generations []uint16
	sparse      []int32
}

// NewPool reserves space for capacity components.
// Nonpositive capacities select a small default allocation.
func NewPool[T any](capacity int) *Pool[T] {
	pool := &Pool[T]{}
	pool.initialize(storageCapacity(capacity))
	return pool
}

func (pool *Pool[T]) initialize(capacity int) {
	if pool == nil || pool.sparse != nil {
		return
	}
	pool.components = make([]T, 0, capacity)
	pool.entities = make([]uint32, 0, capacity)
	pool.generations = make([]uint16, 0, capacity)
	pool.sparse = make([]int32, capacity)
	for index := range pool.sparse {
		pool.sparse[index] = -1
	}
}

func (pool *Pool[T]) ensureSparseCapacity(id uint32) {
	if uint64(id) < uint64(len(pool.sparse)) {
		return
	}
	size := uint64(len(pool.sparse)) * 2
	if required := uint64(id) + 1; size < required {
		size = required
	}
	sparse := make([]int32, size)
	copy(sparse, pool.sparse)
	for index := len(pool.sparse); index < len(sparse); index++ {
		sparse[index] = -1
	}
	pool.sparse = sparse
}

// Add inserts or replaces the component for an ID, recording its generation.
// The caller must pass a live entity from the owning EntityManager; Pool cannot
// independently validate entity liveness. A nil receiver is a no-op.
func (pool *Pool[T]) Add(entity Entity, comp T) {
	if pool == nil {
		return
	}
	pool.initialize(defaultCapacity)
	pool.ensureSparseCapacity(entity.ID)
	if index := pool.sparse[entity.ID]; index != -1 {
		pool.components[index] = comp
		pool.generations[index] = entity.Generation
		return
	}
	if uint64(len(pool.components)) > uint64(^uint32(0)>>1) {
		panic("myth: component index space exhausted")
	}
	pool.sparse[entity.ID] = int32(len(pool.components))
	pool.components = append(pool.components, comp)
	pool.entities = append(pool.entities, entity.ID)
	pool.generations = append(pool.generations, entity.Generation)
}

// Get returns nil for a missing ID, a stale generation, or a nil receiver.
// The returned pointer must not be retained across Add or Remove calls.
func (pool *Pool[T]) Get(entity Entity) *T {
	if pool == nil || uint64(entity.ID) >= uint64(len(pool.sparse)) {
		return nil
	}
	index := pool.sparse[entity.ID]
	if index == -1 || pool.generations[index] != entity.Generation {
		return nil
	}
	return &pool.components[index]
}

// Has reports whether the ID and generation match a stored component.
// Missing entities, stale generations, and nil receivers return false.
func (pool *Pool[T]) Has(entity Entity) bool {
	if pool == nil || uint64(entity.ID) >= uint64(len(pool.sparse)) {
		return false
	}
	index := pool.sparse[entity.ID]
	return index != -1 && pool.generations[index] == entity.Generation
}

// Clear removes all components while retaining the backing arrays for reuse.
// Component references remain in the backing array until overwritten or the
// pool is released. A nil receiver is a no-op.
func (pool *Pool[T]) Clear() {
	if pool == nil {
		return
	}
	pool.components = pool.components[:0]
	pool.entities = pool.entities[:0]
	pool.generations = pool.generations[:0]
	for index := range pool.sparse {
		pool.sparse[index] = -1
	}
}

// Remove swap-removes a matching entity. Missing or stale entities and nil
// receivers are ignored. Component iteration order is not preserved.
func (pool *Pool[T]) Remove(entity Entity) {
	if pool == nil || uint64(entity.ID) >= uint64(len(pool.sparse)) {
		return
	}
	index := pool.sparse[entity.ID]
	if index == -1 || pool.generations[index] != entity.Generation {
		return
	}
	last := len(pool.components) - 1
	if int(index) != last {
		pool.components[index] = pool.components[last]
		pool.entities[index] = pool.entities[last]
		pool.generations[index] = pool.generations[last]
		pool.sparse[pool.entities[index]] = index
	}
	var zero T
	pool.components[last] = zero
	pool.components = pool.components[:last]
	pool.entities = pool.entities[:last]
	pool.generations = pool.generations[:last]
	pool.sparse[entity.ID] = -1
}