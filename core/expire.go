package core

import (
	"log"
	"time"
)

func CheckExpire() {

	var sampleSize = 20
	var expiredKeys = 0

	for {

		for key, obj := range store {
			sampleSize--
			
			if sampleSize < 0 {
				break
			}
			
			if obj.ExpiresAt > 0 && obj.ExpiresAt < time.Now().UnixMilli() {
				delete(store, key)
				expiredKeys++
			}
		}

		log.Println(expiredKeys, "keys were expired")
		log.Println("Total keys in store: ", len(store))

		if(float64(expiredKeys) / float64(sampleSize) < 0.25) {
			break
		}
	}
}