package utils

import "sync"

type Container struct {
	Mu        sync.RWMutex
	DataUsage map[string]float64
}
