package limiter

// SkipList represent white list of keys
type SkipList struct {
	Keys []string
}

// HasKey implements basic look-up of given key in SkipList
func (w *SkipList) HasKey(key string) bool {
	for _, k := range w.Keys {
		if k == key {
			return true
		}
	}
	return false
}
