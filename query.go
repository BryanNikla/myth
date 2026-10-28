package myth

// Query2 visits matching entity generations in the smaller pool's dense order.
// Nil pools or callbacks are ignored. Callbacks may update components but must
// not add or remove components from participating pools or retain pointers.
func Query2[A, B any](poolA *Pool[A], poolB *Pool[B], yield func(entity Entity, a *A, b *B)) {
	if poolA == nil || poolB == nil || yield == nil {
		return
	}
	if len(poolA.components) <= len(poolB.components) {
		for index, id := range poolA.entities {
			entity := Entity{ID: id, Generation: poolA.generations[index]}
			if componentB := poolB.Get(entity); componentB != nil {
				yield(entity, &poolA.components[index], componentB)
			}
		}
		return
	}
	for index, id := range poolB.entities {
		entity := Entity{ID: id, Generation: poolB.generations[index]}
		if componentA := poolA.Get(entity); componentA != nil {
			yield(entity, componentA, &poolB.components[index])
		}
	}
}

// Query3 visits matching entity generations in the smallest pool's dense order.
// Nil pools or callbacks are ignored. Callbacks may update components but must
// not add or remove components from participating pools or retain pointers.
func Query3[A, B, C any](poolA *Pool[A], poolB *Pool[B], poolC *Pool[C], yield func(entity Entity, a *A, b *B, c *C)) {
	if poolA == nil || poolB == nil || poolC == nil || yield == nil {
		return
	}
	if len(poolA.components) <= len(poolB.components) && len(poolA.components) <= len(poolC.components) {
		for index, id := range poolA.entities {
			entity := Entity{ID: id, Generation: poolA.generations[index]}
			componentB := poolB.Get(entity)
			if componentB == nil {
				continue
			}
			if componentC := poolC.Get(entity); componentC != nil {
				yield(entity, &poolA.components[index], componentB, componentC)
			}
		}
		return
	}
	if len(poolB.components) <= len(poolC.components) {
		for index, id := range poolB.entities {
			entity := Entity{ID: id, Generation: poolB.generations[index]}
			componentA := poolA.Get(entity)
			if componentA == nil {
				continue
			}
			if componentC := poolC.Get(entity); componentC != nil {
				yield(entity, componentA, &poolB.components[index], componentC)
			}
		}
		return
	}
	for index, id := range poolC.entities {
		entity := Entity{ID: id, Generation: poolC.generations[index]}
		componentA := poolA.Get(entity)
		if componentA == nil {
			continue
		}
		if componentB := poolB.Get(entity); componentB != nil {
			yield(entity, componentA, componentB, &poolC.components[index])
		}
	}
}

// Query4 visits matching entity generations in the smallest pool's dense order.
// Nil pools or callbacks are ignored. Callbacks may update components but must
// not add or remove components from participating pools or retain pointers.
func Query4[A, B, C, D any](poolA *Pool[A], poolB *Pool[B], poolC *Pool[C], poolD *Pool[D], yield func(entity Entity, a *A, b *B, c *C, d *D)) {
	if poolA == nil || poolB == nil || poolC == nil || poolD == nil || yield == nil {
		return
	}
	sizes := [...]int{len(poolA.components), len(poolB.components), len(poolC.components), len(poolD.components)}
	smallest := 0
	for index, size := range sizes {
		if size < sizes[smallest] {
			smallest = index
		}
	}
	switch smallest {
	case 0:
		for index, id := range poolA.entities {
			entity := Entity{ID: id, Generation: poolA.generations[index]}
			componentB := poolB.Get(entity)
			if componentB == nil {
				continue
			}
			componentC := poolC.Get(entity)
			if componentC == nil {
				continue
			}
			componentD := poolD.Get(entity)
			if componentD == nil {
				continue
			}
			yield(entity, &poolA.components[index], componentB, componentC, componentD)
		}
	case 1:
		for index, id := range poolB.entities {
			entity := Entity{ID: id, Generation: poolB.generations[index]}
			componentA := poolA.Get(entity)
			if componentA == nil {
				continue
			}
			componentC := poolC.Get(entity)
			if componentC == nil {
				continue
			}
			componentD := poolD.Get(entity)
			if componentD == nil {
				continue
			}
			yield(entity, componentA, &poolB.components[index], componentC, componentD)
		}
	case 2:
		for index, id := range poolC.entities {
			entity := Entity{ID: id, Generation: poolC.generations[index]}
			componentA := poolA.Get(entity)
			if componentA == nil {
				continue
			}
			componentB := poolB.Get(entity)
			if componentB == nil {
				continue
			}
			componentD := poolD.Get(entity)
			if componentD == nil {
				continue
			}
			yield(entity, componentA, componentB, &poolC.components[index], componentD)
		}
	case 3:
		for index, id := range poolD.entities {
			entity := Entity{ID: id, Generation: poolD.generations[index]}
			componentA := poolA.Get(entity)
			if componentA == nil {
				continue
			}
			componentB := poolB.Get(entity)
			if componentB == nil {
				continue
			}
			componentC := poolC.Get(entity)
			if componentC == nil {
				continue
			}
			yield(entity, componentA, componentB, componentC, &poolD.components[index])
		}
	}
}

// Query5 visits matching entity generations in the smallest pool's dense order.
// Nil pools or callbacks are ignored. Callbacks may update components but must
// not add or remove components from participating pools or retain pointers.
func Query5[A, B, C, D, E any](poolA *Pool[A], poolB *Pool[B], poolC *Pool[C], poolD *Pool[D], poolE *Pool[E], yield func(entity Entity, a *A, b *B, c *C, d *D, e *E)) {
	if poolA == nil || poolB == nil || poolC == nil || poolD == nil || poolE == nil || yield == nil {
		return
	}
	sizes := [...]int{len(poolA.components), len(poolB.components), len(poolC.components), len(poolD.components), len(poolE.components)}
	smallest := 0
	for index, size := range sizes {
		if size < sizes[smallest] {
			smallest = index
		}
	}
	switch smallest {
	case 0:
		for index, id := range poolA.entities {
			entity := Entity{ID: id, Generation: poolA.generations[index]}
			componentB := poolB.Get(entity)
			if componentB == nil {
				continue
			}
			componentC := poolC.Get(entity)
			if componentC == nil {
				continue
			}
			componentD := poolD.Get(entity)
			if componentD == nil {
				continue
			}
			componentE := poolE.Get(entity)
			if componentE == nil {
				continue
			}
			yield(entity, &poolA.components[index], componentB, componentC, componentD, componentE)
		}
	case 1:
		for index, id := range poolB.entities {
			entity := Entity{ID: id, Generation: poolB.generations[index]}
			componentA := poolA.Get(entity)
			if componentA == nil {
				continue
			}
			componentC := poolC.Get(entity)
			if componentC == nil {
				continue
			}
			componentD := poolD.Get(entity)
			if componentD == nil {
				continue
			}
			componentE := poolE.Get(entity)
			if componentE == nil {
				continue
			}
			yield(entity, componentA, &poolB.components[index], componentC, componentD, componentE)
		}
	case 2:
		for index, id := range poolC.entities {
			entity := Entity{ID: id, Generation: poolC.generations[index]}
			componentA := poolA.Get(entity)
			if componentA == nil {
				continue
			}
			componentB := poolB.Get(entity)
			if componentB == nil {
				continue
			}
			componentD := poolD.Get(entity)
			if componentD == nil {
				continue
			}
			componentE := poolE.Get(entity)
			if componentE == nil {
				continue
			}
			yield(entity, componentA, componentB, &poolC.components[index], componentD, componentE)
		}
	case 3:
		for index, id := range poolD.entities {
			entity := Entity{ID: id, Generation: poolD.generations[index]}
			componentA := poolA.Get(entity)
			if componentA == nil {
				continue
			}
			componentB := poolB.Get(entity)
			if componentB == nil {
				continue
			}
			componentC := poolC.Get(entity)
			if componentC == nil {
				continue
			}
			componentE := poolE.Get(entity)
			if componentE == nil {
				continue
			}
			yield(entity, componentA, componentB, componentC, &poolD.components[index], componentE)
		}
	case 4:
		for index, id := range poolE.entities {
			entity := Entity{ID: id, Generation: poolE.generations[index]}
			componentA := poolA.Get(entity)
			if componentA == nil {
				continue
			}
			componentB := poolB.Get(entity)
			if componentB == nil {
				continue
			}
			componentC := poolC.Get(entity)
			if componentC == nil {
				continue
			}
			componentD := poolD.Get(entity)
			if componentD == nil {
				continue
			}
			yield(entity, componentA, componentB, componentC, componentD, &poolE.components[index])
		}
	}
}