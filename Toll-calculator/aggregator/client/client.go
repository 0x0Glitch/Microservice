package client

import (
	"context"

	"github.com/0x0Glitch/toll-calculator/types"
)

type Client interface{
	Aggregate(context.Context,types.AggregatorRequest) error
}