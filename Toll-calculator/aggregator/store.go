package main

import (
	"fmt"

	"github.com/0x0Glitch/toll-calculator/types"
)


type MemoryStore struct{
	data map[int]float64
}

func NewMemoryStore() *MemoryStore{
	return &MemoryStore{
		data:make(map[int]float64),
	}
}

func (m *MemoryStore) Insert(d types.Distance) error{
	m.data[d.OBUID] += d.Values
	return nil
}

func (m *MemoryStore) Get(id int) (float64,error){
	dist,ok := m.data[id]
	if !ok{
		return 0.0,fmt.Errorf("Couldn't find distance for id: %d",id)
	}
	return dist,nil
}