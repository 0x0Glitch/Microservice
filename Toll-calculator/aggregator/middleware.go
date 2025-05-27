package main

import (
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct{
	next Aggregator
}


func NewLogMiddleware(next Aggregator) Aggregator{
	return &LogMiddleware{
		next:next,
	}
}
func (m *LogMiddleware) AggregateDistance(distance types.Distance) (err error){
	defer func(start time.Time){
		logrus.WithFields(logrus.Fields{
			"time":time.Since(start),
			"err":err,
		}).Info("Aggregate distance")
	}(time.Now())
	err = m.next.AggregateDistance(distance)
	return
}