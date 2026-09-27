package core

type RedisCmd struct {
	Command string
	Args    []string
}

type RedisCmds []*RedisCmd