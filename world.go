package myth

// UntypedPool provides lifecycle operations shared by all component pools.
type UntypedPool interface {
	Remove(entity Entity)
	Clear()
}

// World coordinates entity allocation and registered component pool cleanup.
// Its zero value is usable. It must not be copied after first use, and its
// operations must be externally serialized.
type World struct {
	manager *EntityManager
	pools   []UntypedPool
}

// NewWorld reserves capacity for entity metadata and prepares pool tracking.
func NewWorld(capacity int) *World {
	return &World{
		manager: NewEntityManager(capacity),
		pools:   make([]UntypedPool, 0),
	}
}

// RegisterPool registers a pool for destruction and clearing.
// Nil receivers and nil interfaces are ignored.
func (world *World) RegisterPool(pool UntypedPool) {
	if world == nil || pool == nil {
		return
	}
	world.pools = append(world.pools, pool)
}

// Create allocates a live entity. A nil receiver returns the invalid zero Entity.
func (world *World) Create() Entity {
	if world == nil {
		return Entity{}
	}
	if world.manager == nil {
		world.manager = NewEntityManager(defaultCapacity)
	}
	return world.manager.Create()
}

// Destroy removes matching components from registered pools before recycling
// the entity ID. A nil receiver is a no-op.
func (world *World) Destroy(entity Entity) {
	if world == nil {
		return
	}
	for _, pool := range world.pools {
		pool.Remove(entity)
	}
	world.manager.Destroy(entity)
}

// Clear clears registered pools without destroying entities or recycling IDs.
// A nil receiver is a no-op.
func (world *World) Clear() {
	if world == nil {
		return
	}
	for _, pool := range world.pools {
		pool.Clear()
	}
}