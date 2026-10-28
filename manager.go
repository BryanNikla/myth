package myth

// EntityManager allocates entity IDs and recycles dead IDs in LIFO order.
// Its zero value is usable. It must not be copied after first use, and its
// operations must be externally serialized. It does not own component pools.
type EntityManager struct {
	generations []uint16
	alive       []bool
	freeIDs     []uint32
}

// NewEntityManager reserves capacity for entity metadata and recycled IDs.
func NewEntityManager(capacity int) *EntityManager {
	manager := &EntityManager{}
	manager.initialize(storageCapacity(capacity))
	return manager
}

func (manager *EntityManager) initialize(capacity int) {
	if manager == nil || manager.generations != nil {
		return
	}
	manager.generations = make([]uint16, 0, capacity)
	manager.alive = make([]bool, 0, capacity)
	manager.freeIDs = make([]uint32, 0, capacity)
}

// Create allocates a live entity. A nil receiver returns the invalid zero
// Entity. It panics if all 32-bit IDs have been allocated or retired.
func (manager *EntityManager) Create() Entity {
	if manager == nil {
		return Entity{}
	}
	manager.initialize(defaultCapacity)
	if len(manager.freeIDs) > 0 {
		last := len(manager.freeIDs) - 1
		id := manager.freeIDs[last]
		manager.freeIDs = manager.freeIDs[:last]
		manager.generations[id]++
		manager.alive[id] = true
		return Entity{ID: id, Generation: manager.generations[id]}
	}
	if uint64(len(manager.generations)) > uint64(^uint32(0)) {
		panic("myth: entity ID space exhausted")
	}
	id := uint32(len(manager.generations))
	manager.generations = append(manager.generations, 1)
	manager.alive = append(manager.alive, true)
	return Entity{ID: id, Generation: 1}
}

// IsAlive checks both liveness and generation. Nil receivers return false.
func (manager *EntityManager) IsAlive(entity Entity) bool {
	if manager == nil || uint64(entity.ID) >= uint64(len(manager.generations)) {
		return false
	}
	return manager.alive[entity.ID] && manager.generations[entity.ID] == entity.Generation
}

// Destroy ignores stale, unknown, and already-dead entities. Remove the
// entity from its component pools separately. IDs at the maximum generation
// are retired permanently so generation wraparound cannot revive old handles.
func (manager *EntityManager) Destroy(entity Entity) {
	if manager == nil || !manager.IsAlive(entity) {
		return
	}
	manager.alive[entity.ID] = false
	if entity.Generation != ^uint16(0) {
		manager.freeIDs = append(manager.freeIDs, entity.ID)
	}
}