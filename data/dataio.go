package data

// Save writes the persistable entity to its canonical location using the default store backend.
func Save(ps Persistable) error {
	return defaultStore.save(ps)
}

// Update overwrites the stored representation even if it already exists on disk.
func Update(ps Persistable) error {
	return defaultStore.update(ps)
}

// Persisted reports whether the entity already has on-disk state without surfacing lookup errors.
func Persisted(ps Persistable) bool {
	persisted, err := defaultStore.persisted(ps)
	return err == nil && persisted
}

// Load hydrates the entity from its location, leaving validation to the caller.
func Load(ps Persistable) error {
	return defaultStore.load(ps)
}

// LoadFromPath hydrates an entity from an explicit path, bypassing Location-derived lookup.
func LoadFromPath(path string, ps Persistable) error {
	return defaultStore.loadFromPath(path, ps)
}

// Destroy removes the entity's persisted form and runs its cleanup hook.
func Destroy(ps Destroyable) error {
	return defaultStore.destroy(ps)
}
