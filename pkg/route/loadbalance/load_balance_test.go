package loadbalance

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"knoway.dev/api/route/v1alpha1"
)

func TestWeightedRoundRobin_Next(t *testing.T) {
	destinations := []*v1alpha1.RouteDestination{
		{
			Cluster:   "backend1",
			Namespace: "namespace1",
			Weight:    lo.ToPtr(int32(10)),
		},
		{
			Cluster:   "backend2",
			Namespace: "namespace2",
			Weight:    lo.ToPtr(int32(20)),
		},
		{
			Cluster:   "backend3",
			Namespace: "namespace3",
			Weight:    lo.ToPtr(int32(50)),
		},
		{
			Cluster:   "backend4",
			Namespace: "namespace4",
			Weight:    lo.ToPtr(int32(20)),
		},
	}

	lb := NewWeightedRoundRobin(destinations)

	var (
		numBackend1, numBackend2, numBackend3, numBackend4 atomic.Int32
	)

	wg := &sync.WaitGroup{}
	for range 20 {
		wg.Add(1)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			for range 10000 {
				next := lb.Next(context.TODO(), nil)
				switch next {
				case "backend1":
					numBackend1.Add(1)
				case "backend2":
					numBackend2.Add(1)
				case "backend3":
					numBackend3.Add(1)
				case "backend4":
					numBackend4.Add(1)
				default:
					panic("unknown backend")
				}
			}
		}(wg)
	}

	wg.Wait()

	total := float32(numBackend1.Load() + numBackend2.Load() + numBackend3.Load() + numBackend4.Load())
	threshold1 := float32(numBackend3.Load())
	lowerBound1 := threshold1 * 0.99
	upperBound1 := threshold1 * 1.01
	expect1 := lowerBound1/total <= float32(numBackend3.Load())/total && float32(numBackend3.Load())/total <= upperBound1/total
	assert.True(t, expect1)

	threshold2 := float32(numBackend2.Load())
	lowerBound2 := threshold2 * 0.99
	upperBound2 := threshold2 * 1.01
	expect2 := (lowerBound2/total) <= (float32(numBackend2.Load())/total) && (float32(numBackend2.Load())/total) <= upperBound2/total
	assert.True(t, expect2)

	threshold3 := float32(numBackend4.Load())
	lowerBound3 := threshold3 * 0.99
	upperBound3 := threshold3 * 1.01
	expect3 := (lowerBound3/total) <= (float32(numBackend4.Load())/total) && (float32(numBackend4.Load())/total) <= (upperBound3/total)
	assert.True(t, expect3)
}

func TestWeightedLeastRequest_Next(t *testing.T) {
	destinations := []*v1alpha1.RouteDestination{
		{
			Cluster:   "backend1",
			Namespace: "namespace1",
			Weight:    lo.ToPtr(int32(10)),
		},
		{
			Cluster:   "backend2",
			Namespace: "namespace2",
			Weight:    lo.ToPtr(int32(60)),
		},
		{
			Cluster:   "backend3",
			Namespace: "namespace3",
			Weight:    lo.ToPtr(int32(30)),
		},
	}

	lb := NewWeightedLeastRequest(destinations)

	var (
		numBackend1, numBackend2, numBackend3 int
	)

	for range 100 {
		defer lb.Done(context.TODO())
		bak := lb.Next(context.TODO(), nil)

		switch bak {
		case "backend1":
			numBackend1++
		case "backend2":
			numBackend2++
		case "backend3":
			numBackend3++
		default:
			panic("unknown backend")
		}
	}

	total := float32(numBackend1 + numBackend2 + numBackend3)
	threshold1 := float32(numBackend3)
	lowerBound1 := threshold1 * 0.9
	upperBound1 := threshold1 * 1.1
	expect1 := lowerBound1 <= (float32(numBackend3)/total)*100 && (float32(numBackend3)/total)*100 <= upperBound1
	assert.True(t, expect1)

	threshold2 := float32(numBackend2)
	lowerBound2 := threshold2 * 0.9
	upperBound2 := threshold2 * 1.1
	expect2 := lowerBound2 <= (float32(numBackend2)/total)*100 && (float32(numBackend2)/total)*100 <= upperBound2
	assert.True(t, expect2)

	threshold3 := float32(numBackend1)
	lowerBound3 := threshold3 * 0.9
	upperBound3 := threshold3 * 1.1
	expect3 := lowerBound3 <= (float32(numBackend1)/total)*100 && (float32(numBackend1)/total)*100 <= upperBound3
	assert.True(t, expect3)
}
