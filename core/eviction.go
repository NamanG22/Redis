package core

func evictFirst() {
	for key, _ := range store {
		delete(store, key)
		return
	}
}

func evict() {
	// can implement various complex algos, but choosing a simple algo 
	evictFirst()
}