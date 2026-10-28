package myth

import "testing"

func TestPoolLifecycle(t *testing.T) {
	pool := NewPool[int](2)
	first := Entity{ID: 10, Generation: 1}
	second := Entity{ID: 20, Generation: 3}
	third := Entity{ID: 1000, Generation: 2}
	if pool.Get(first) != nil {
		t.Fatal("missing entity returned a component")
	}
	pool.Add(first, 10)
	pool.Add(second, 20)
	pool.Add(third, 30)
	pool.Add(first, 11)
	if len(pool.components) != 3 || *pool.Get(first) != 11 {
		t.Fatal("replacement changed pool length or failed to update")
	}
	*pool.Get(second) = 21
	stale := Entity{ID: second.ID, Generation: second.Generation - 1}
	pool.Remove(stale)
	if pool.Get(stale) != nil || *pool.Get(second) != 21 {
		t.Fatal("stale handle accessed or removed a component")
	}
	pool.Remove(first)
	if pool.Get(first) != nil || *pool.Get(third) != 30 {
		t.Fatal("swap removal failed to repair sparse index")
	}
	pool.Remove(first)
	pool.Remove(second)
	pool.Remove(third)
	if len(pool.components) != 0 {
		t.Fatal("pool did not empty")
	}
	for _, index := range pool.sparse {
		if index != -1 {
			t.Fatal("empty sparse slot did not contain -1")
		}
	}
}

func TestPoolGenerationReplacementAndZeroValue(t *testing.T) {
	var pool Pool[int]
	old := Entity{ID: 7, Generation: 1}
	current := Entity{ID: 7, Generation: 2}
	pool.Remove(old)
	pool.Add(old, 1)
	pool.Add(current, 2)
	pool.Remove(old)
	if pool.Get(old) != nil || *pool.Get(current) != 2 || len(pool.components) != 1 {
		t.Fatal("generation replacement failed")
	}
}

func TestPoolRemoveClearsReferences(t *testing.T) {
	pool := NewPool[*int](2)
	value := 1
	first := Entity{ID: 1}
	second := Entity{ID: 2}
	pool.Add(first, &value)
	pool.Add(second, &value)
	pool.Remove(first)
	if pool.components[:cap(pool.components)][1] != nil {
		t.Fatal("removed dense slot retained a reference")
	}
}

func TestManagerLifecycle(t *testing.T) {
	var manager EntityManager
	first := manager.Create()
	second := manager.Create()
	if first.ID == second.ID || !manager.IsAlive(first) || manager.IsAlive(Entity{}) {
		t.Fatal("invalid initial entity state")
	}
	manager.Destroy(first)
	manager.Destroy(first)
	manager.Destroy(Entity{ID: ^uint32(0), Generation: 1})
	if manager.IsAlive(first) || len(manager.freeIDs) != 1 {
		t.Fatal("destroy failed or double destroy corrupted free list")
	}
	recycled := manager.Create()
	manager.Destroy(first)
	if recycled.ID != first.ID || recycled.Generation != first.Generation+1 {
		t.Fatal("ID was not recycled with a new generation")
	}
	if !manager.IsAlive(recycled) || manager.IsAlive(first) {
		t.Fatal("stale destroy affected the recycled entity")
	}
	third := manager.Create()
	if third.ID == recycled.ID || third.ID == second.ID {
		t.Fatal("live ID allocated twice")
	}
}

func TestManagerRetiresExhaustedGenerations(t *testing.T) {
	manager := NewEntityManager(1)
	entity := manager.Create()
	manager.generations[entity.ID] = ^uint16(0) - 1
	entity.Generation = ^uint16(0) - 1
	manager.Destroy(entity)
	last := manager.Create()
	if last.Generation != ^uint16(0) {
		t.Fatal("last usable generation was skipped")
	}
	manager.Destroy(last)
	manager.Destroy(last)
	next := manager.Create()
	if next.ID == last.ID || manager.IsAlive(last) || len(manager.freeIDs) != 0 {
		t.Fatal("exhausted ID was recycled")
	}
}

func TestNilReceivers(t *testing.T) {
	var pool *Pool[int]
	var manager *EntityManager
	entity := Entity{ID: 1, Generation: 1}
	pool.Add(entity, 1)
	pool.Remove(entity)
	manager.Destroy(entity)
	if pool.Get(entity) != nil || manager.IsAlive(entity) || manager.Create() != (Entity{}) {
		t.Fatal("unexpected nil receiver result")
	}
}

func TestRecycledEntityAcrossPool(t *testing.T) {
	manager := NewEntityManager(1)
	pool := NewPool[int](1)
	old := manager.Create()
	pool.Add(old, 10)
	pool.Remove(old)
	manager.Destroy(old)
	current := manager.Create()
	pool.Add(current, 20)
	pool.Remove(old)
	manager.Destroy(old)
	if pool.Get(old) != nil || *pool.Get(current) != 20 || !manager.IsAlive(current) {
		t.Fatal("stale handle affected a recycled entity")
	}
}

func TestPoolSparseGrowth(t *testing.T) {
	pool := NewPool[int](2)
	first := Entity{ID: 1, Generation: 1}
	pool.Add(first, 10)
	for _, id := range []uint32{2, 1000} {
		entity := Entity{ID: id, Generation: 1}
		pool.Add(entity, 20)
		if *pool.Get(first) != 10 || *pool.Get(entity) != 20 {
			t.Fatal("sparse growth lost existing indices")
		}
	}
	for _, id := range []uint32{0, 3, 999, ^uint32(0)} {
		entity := Entity{ID: id, Generation: 1}
		if pool.Get(entity) != nil {
			t.Fatalf("missing ID %d returned a component", id)
		}
		pool.Remove(entity)
	}
	pool.Remove(first)
	pool.Add(first, 30)
	if *pool.Get(first) != 30 || len(pool.components) != 3 {
		t.Fatal("removed sparse slot could not be reused")
	}
}

func TestQuery2Drivers(t *testing.T) {
	for _, driver := range []int{0, 1, 2} {
		t.Run([]string{"A", "B", "tie"}[driver], func(t *testing.T) {
			poolA := NewPool[int](1)
			poolB := NewPool[string](1)
			first := Entity{ID: 3, Generation: 2}
			second := Entity{ID: 7, Generation: 2}
			poolA.Add(first, 10)
			poolA.Add(second, 20)
			poolB.Add(second, "second")
			poolB.Add(first, "first")
			poolA.Add(Entity{ID: 9, Generation: 1}, 0)
			poolB.Add(Entity{ID: 9, Generation: 2}, "stale")
			poolA.Add(Entity{ID: 20, Generation: 1}, 0)
			poolB.Add(Entity{ID: 21, Generation: 1}, "missing")
			if driver == 0 {
				poolB.Add(Entity{ID: 30, Generation: 1}, "padding")
			}
			if driver == 1 {
				poolA.Add(Entity{ID: 30, Generation: 1}, 0)
			}
			want := []Entity{first, second}
			if driver == 1 {
				want = []Entity{second, first}
			}
			count := 0
			Query2(poolA, poolB, func(entity Entity, componentA *int, componentB *string) {
				if count >= len(want) || entity != want[count] {
					t.Fatalf("unexpected entity or driver order: %v", entity)
				}
				if componentA != poolA.Get(entity) || componentB != poolB.Get(entity) {
					t.Fatal("callback did not receive stored component pointers")
				}
				*componentA += 1
				*componentB += "!"
				count++
			})
			if count != 2 || *poolA.Get(first) != 11 || *poolA.Get(second) != 21 ||
				*poolB.Get(first) != "first!" || *poolB.Get(second) != "second!" {
				t.Fatal("query missed matches or failed to mutate components")
			}
		})
	}
}

func TestQuery3Drivers(t *testing.T) {
	for _, driver := range []int{0, 1, 2, 3} {
		t.Run([]string{"A", "B", "C", "tie"}[driver], func(t *testing.T) {
			poolA := NewPool[int](1)
			poolB := NewPool[string](1)
			poolC := NewPool[float64](1)
			first := Entity{ID: 3, Generation: 2}
			second := Entity{ID: 7, Generation: 2}
			poolA.Add(first, 10)
			poolA.Add(second, 20)
			poolB.Add(second, "second")
			poolB.Add(first, "first")
			poolC.Add(second, 2.5)
			poolC.Add(first, 1.5)
			poolA.Add(Entity{ID: 9, Generation: 1}, 0)
			poolB.Add(Entity{ID: 9, Generation: 2}, "stale")
			poolC.Add(Entity{ID: 9, Generation: 3}, 0)
			for missing := 0; missing < 3; missing++ {
				entity := Entity{ID: uint32(20 + missing), Generation: 1}
				if missing != 0 {
					poolA.Add(entity, 0)
				}
				if missing != 1 {
					poolB.Add(entity, "missing")
				}
				if missing != 2 {
					poolC.Add(entity, 0)
				}
			}
			padding := Entity{ID: 30, Generation: 1}
			if driver != 0 && driver != 3 {
				poolA.Add(padding, 0)
			}
			if driver != 1 && driver != 3 {
				poolB.Add(padding, "padding")
			}
			if driver != 2 && driver != 3 {
				poolC.Add(padding, 0)
			}
			want := []Entity{first, second}
			if driver == 1 || driver == 2 {
				want = []Entity{second, first}
			}
			count := 0
			Query3(poolA, poolB, poolC, func(entity Entity, componentA *int, componentB *string, componentC *float64) {
				if count >= len(want) || entity != want[count] {
					t.Fatalf("unexpected entity or driver order: %v", entity)
				}
				if componentA != poolA.Get(entity) || componentB != poolB.Get(entity) || componentC != poolC.Get(entity) {
					t.Fatal("callback did not receive stored component pointers")
				}
				*componentA += 1
				*componentB += "!"
				*componentC += 1
				count++
			})
			if count != 2 || *poolA.Get(first) != 11 || *poolA.Get(second) != 21 ||
				*poolB.Get(first) != "first!" || *poolB.Get(second) != "second!" ||
				*poolC.Get(first) != 2.5 || *poolC.Get(second) != 3.5 {
				t.Fatal("query missed matches or failed to mutate components")
			}
		})
	}
}

func TestQueriesNilAndEmpty(t *testing.T) {
	var nilPool *Pool[int]
	var empty Pool[int]
	populated := NewPool[int](1)
	populated.Add(Entity{ID: 1, Generation: 1}, 10)
	yield2 := func(Entity, *int, *int) { t.Fatal("unexpected query match") }
	yield3 := func(Entity, *int, *int, *int) { t.Fatal("unexpected query match") }
	for _, pool := range []*Pool[int]{nilPool, &empty} {
		Query2(pool, populated, yield2)
		Query2(populated, pool, yield2)
		Query3(pool, populated, populated, yield3)
		Query3(populated, pool, populated, yield3)
		Query3(populated, populated, pool, yield3)
	}
	Query2(populated, populated, nil)
	Query3(populated, populated, populated, nil)
	yield4 := func(Entity, *int, *int, *int, *int) { t.Fatal("unexpected query match") }
	yield5 := func(Entity, *int, *int, *int, *int, *int) { t.Fatal("unexpected query match") }
	for _, pool := range []*Pool[int]{nilPool, &empty} {
		for position := 0; position < 5; position++ {
			pools := [5]*Pool[int]{populated, populated, populated, populated, populated}
			pools[position] = pool
			if position < 4 {
				Query4(pools[0], pools[1], pools[2], pools[3], yield4)
			}
			Query5(pools[0], pools[1], pools[2], pools[3], pools[4], yield5)
		}
	}
	Query4(populated, populated, populated, populated, nil)
	Query5(populated, populated, populated, populated, populated, nil)
}

func TestQuery4Drivers(t *testing.T) {
	testQueryDrivers(t, 4, func(pools []*Pool[int], yield func(Entity, []*int)) {
		Query4(pools[0], pools[1], pools[2], pools[3], func(entity Entity, componentA, componentB, componentC, componentD *int) {
			yield(entity, []*int{componentA, componentB, componentC, componentD})
		})
	})
}

func TestQuery5Drivers(t *testing.T) {
	testQueryDrivers(t, 5, func(pools []*Pool[int], yield func(Entity, []*int)) {
		Query5(pools[0], pools[1], pools[2], pools[3], pools[4], func(entity Entity, componentA, componentB, componentC, componentD, componentE *int) {
			yield(entity, []*int{componentA, componentB, componentC, componentD, componentE})
		})
	})
}

func TestPoolHasAndClear(t *testing.T) {
	var nilPool *Pool[int]
	var pool Pool[int]
	entity := Entity{ID: 100, Generation: 2}
	nilPool.Clear()
	pool.Clear()
	if nilPool.Has(entity) || pool.Has(entity) {
		t.Fatal("nil or empty pool reported a component")
	}
	pool.Add(entity, 10)
	for _, candidate := range []Entity{
		entity,
		{ID: entity.ID, Generation: 1},
		{ID: 0, Generation: 2},
		{ID: ^uint32(0), Generation: 2},
	} {
		if pool.Has(candidate) != (pool.Get(candidate) != nil) {
			t.Fatalf("Has disagrees with Get for %v", candidate)
		}
	}
	components := &pool.components[0]
	entities := &pool.entities[0]
	generations := &pool.generations[0]
	sparse := &pool.sparse[0]
	pool.Clear()
	if len(pool.components) != 0 || len(pool.entities) != 0 || len(pool.generations) != 0 || pool.Has(entity) {
		t.Fatal("Clear left dense data or a component visible")
	}
	for _, index := range pool.sparse {
		if index != -1 {
			t.Fatal("Clear left an occupied sparse slot")
		}
	}
	pool.Clear()
	pool.Add(entity, 20)
	if &pool.components[0] != components || &pool.entities[0] != entities ||
		&pool.generations[0] != generations || &pool.sparse[0] != sparse {
		t.Fatal("Clear did not retain backing arrays")
	}
	if !pool.Has(entity) || *pool.Get(entity) != 20 {
		t.Fatal("pool could not be reused after Clear")
	}
	pool.Remove(entity)
	if pool.Has(entity) {
		t.Fatal("removed entity still has a component")
	}
}

func TestWorldLifecycle(t *testing.T) {
	world := NewWorld(2)
	positions := NewPool[int](2)
	velocities := NewPool[float64](2)
	unregistered := NewPool[string](2)
	world.RegisterPool(positions)
	world.RegisterPool(velocities)
	entity := world.Create()
	positions.Add(entity, 10)
	velocities.Add(entity, 2.5)
	unregistered.Add(entity, "unregistered")
	world.Destroy(entity)
	world.Destroy(entity)
	if positions.Has(entity) || velocities.Has(entity) || world.manager.IsAlive(entity) {
		t.Fatal("Destroy did not remove components and retire the entity")
	}
	if !unregistered.Has(entity) {
		t.Fatal("Destroy affected an unregistered pool")
	}
	recycled := world.Create()
	if recycled.ID != entity.ID || recycled.Generation != entity.Generation+1 {
		t.Fatal("Destroy did not recycle the entity ID")
	}
	positions.Add(recycled, 20)
	velocities.Add(recycled, 5)
	world.Destroy(entity)
	if !positions.Has(recycled) || !velocities.Has(recycled) || !world.manager.IsAlive(recycled) {
		t.Fatal("stale destruction affected the recycled entity")
	}
	world.Clear()
	world.Clear()
	if positions.Has(recycled) || velocities.Has(recycled) || !world.manager.IsAlive(recycled) {
		t.Fatal("Clear failed to clear components or destroyed a live entity")
	}
	positions.Add(recycled, 30)
	if !positions.Has(recycled) || world.Create().ID == recycled.ID {
		t.Fatal("Clear prevented reuse or recycled a live ID")
	}
}

func TestWorldNilAndZeroValue(t *testing.T) {
	var nilWorld *World
	nilWorld.RegisterPool(NewPool[int](1))
	nilWorld.Destroy(Entity{})
	nilWorld.Clear()
	if nilWorld.Create() != (Entity{}) {
		t.Fatal("nil world returned a live entity")
	}
	var world World
	world.RegisterPool(nil)
	if len(world.pools) != 0 {
		t.Fatal("nil interface was registered")
	}
	var nilPool *Pool[int]
	world.RegisterPool(nilPool)
	world.Destroy(Entity{})
	world.Clear()
	entity := world.Create()
	if !world.manager.IsAlive(entity) {
		t.Fatal("zero-value world failed to create a live entity")
	}
	world.Destroy(entity)
	if world.manager.IsAlive(entity) {
		t.Fatal("zero-value world failed to destroy an entity")
	}
}

func testQueryDrivers(t *testing.T, poolCount int, query func([]*Pool[int], func(Entity, []*int))) {
	t.Helper()
	for driver := 0; driver <= poolCount; driver++ {
		name := "tie"
		if driver < poolCount {
			name = []string{"A", "B", "C", "D", "E"}[driver]
		}
		t.Run(name, func(t *testing.T) {
			selected := driver
			if selected == poolCount {
				selected = 0
			}
			first := Entity{ID: 3, Generation: 2}
			second := Entity{ID: 7, Generation: 2}
			pools := make([]*Pool[int], poolCount)
			for poolIndex := range pools {
				pool := NewPool[int](1)
				pools[poolIndex] = pool
				if poolIndex == selected {
					pool.Add(second, 20+poolIndex)
					pool.Add(first, 10+poolIndex)
				} else {
					pool.Add(first, 10+poolIndex)
					pool.Add(second, 20+poolIndex)
				}
				for excluded := 0; excluded < poolCount; excluded++ {
					if poolIndex != excluded {
						pool.Add(Entity{ID: uint32(20 + excluded), Generation: 1}, 0)
					}
					stale := Entity{ID: uint32(40 + excluded), Generation: 1}
					if poolIndex == excluded {
						stale.Generation = 2
					}
					pool.Add(stale, 0)
				}
				if driver < poolCount && poolIndex != selected {
					pool.Add(Entity{ID: uint32(100 + poolIndex), Generation: 1}, 0)
				}
			}
			want := []Entity{second, first}
			count := 0
			query(pools, func(entity Entity, components []*int) {
				if count >= len(want) || entity != want[count] {
					t.Fatalf("unexpected entity or driver order: %v", entity)
				}
				for poolIndex, component := range components {
					if component != pools[poolIndex].Get(entity) {
						t.Fatalf("component %d was not the stored pointer", poolIndex)
					}
					*component += 1
				}
				count++
			})
			if count != len(want) {
				t.Fatalf("got %d matches, want %d", count, len(want))
			}
			for poolIndex, pool := range pools {
				if *pool.Get(first) != 11+poolIndex || *pool.Get(second) != 21+poolIndex {
					t.Fatalf("pool %d components were not updated", poolIndex)
				}
			}
		})
	}
}