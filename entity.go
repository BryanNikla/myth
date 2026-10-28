// Package myth provides single-owner, mutex-free ECS storage.
// It does not provide synchronization for concurrent readers and writers.
package myth

// Entity identifies an entity incarnation within one EntityManager.
type Entity struct {
	ID         uint32
	Generation uint16
}

const defaultCapacity = 64

func storageCapacity(capacity int) int {
	if capacity <= 0 {
		return defaultCapacity
	}
	return capacity
}