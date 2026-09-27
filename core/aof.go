package core

import (
	"fmt"
	"log"
	"os"

	"github.com/NamanG22/Redis/config"
)

func DumpAllAOF() error {
	tmpFile := config.AOFFile + ".tmp"
	fp, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("error rewriting AOF:", err)
		return err
	}

	log.Println("Rewriting AOF file at ", config.AOFFile)
	for key := range store {
		obj := Get(key)
		if obj == nil {
			continue
		}
		if err := dumpKey(fp, key, obj); err != nil {
			fp.Close()
			os.Remove(tmpFile)
			return err
		}
	}

	if err := fp.Close(); err != nil {
		os.Remove(tmpFile)
		return err
	}
	if err := os.Rename(tmpFile, config.AOFFile); err != nil {
		os.Remove(tmpFile)
		return err
	}
	log.Println("AOF file rewritten successfully")
	return nil
}

func dumpKey(fp *os.File, key string, val *Obj) error {
	value, ok := val.Value.(string)
	if !ok {
		value = fmt.Sprint(val.Value)
	}
	_, err := fp.Write(Encode([]string{"SET", key, value}, false))
	return err
}
