package loadbalance

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
	"sync/atomic"

	"github.com/samber/lo"

	"knoway.dev/api/route/v1alpha1"
	"knoway.dev/pkg/object"
)

type LoadBalancer interface {
	// Next returns the next destination to send the request to.
	Next(ctx context.Context, request object.LLMRequest) string
	Done(ctx context.Context)
}

type server struct {
	name           string
	weight         int32
	requestCounter requestCounter
}

type requestCounter interface {
	Current() int
	Inc()
	Desc()
	Less(o requestCounter) bool
}

func newServers(destinations []*v1alpha1.RouteDestination) []*server {
	servers := make([]*server, 0)
	for _, d := range destinations {
		servers = append(servers, &server{
			name:   d.GetCluster(),
			weight: d.GetWeight(),
			// TODO: if knoway deploy in mutil replicas, should implement a distributed request counter
			requestCounter: newRequestCounter(),
		})
	}

	return servers
}

func newRequestCounter() requestCounter {
	return &memoryRequestCounter{
		count: atomic.Int32{},
	}
}

var _ LoadBalancer = (*WeightedRoundRobin)(nil)

type WeightedRoundRobin struct {
	servers     []*server
	current     atomic.Int32
	totalWeight int
}

func NewWeightedRoundRobin(destinations []*v1alpha1.RouteDestination) *WeightedRoundRobin {
	return &WeightedRoundRobin{
		servers: newServers(destinations),
		totalWeight: lo.SumBy(destinations, func(item *v1alpha1.RouteDestination) int {
			return int(item.GetWeight())
		}),
	}
}

func (w *WeightedRoundRobin) calculateTotalWeight() int64 {
	return lo.SumBy(w.servers, func(item *server) int64 {
		return int64(item.weight)
	})
}

func (w *WeightedRoundRobin) Done(_ context.Context) {}

func (w *WeightedRoundRobin) Next(ctx context.Context, _ object.LLMRequest) string {
	if len(w.servers) == 0 {
		return ""
	}
	if len(w.servers) == 1 {
		return w.servers[0].name
	}

	knownTotalWeight := w.calculateTotalWeight()

	randomWeight, err := rand.Int(rand.Reader, big.NewInt(knownTotalWeight))
	if err != nil {
		return ""
	}

	currentIndex := w.current.Load()
	var currentWeight int32
	var total int64
	foundIdx := -1

	for i := range w.servers {
		idx := (int(currentIndex) + i) % len(w.servers)
		currentWeight = w.servers[idx].weight
		total += int64(currentWeight)

		if total > randomWeight.Int64() {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		foundIdx = int(currentIndex)
	}

	nextIndex := (int32(foundIdx) + 1) % int32(len(w.servers))
	w.current.Store(nextIndex)
	selectedService := w.servers[foundIdx]

	return selectedService.name
}

var _ LoadBalancer = (*WeightedLeastRequest)(nil)

type WeightedLeastRequest struct {
	servers []*server
	current int
}

func NewWeightedLeastRequest(destinations []*v1alpha1.RouteDestination) LoadBalancer {
	return &WeightedLeastRequest{
		servers: newServers(destinations),
	}
}

type memoryRequestCounter struct {
	count atomic.Int32
}

func (m *memoryRequestCounter) Less(o requestCounter) bool {
	return m.Current() < o.Current()
}

func (m *memoryRequestCounter) Current() int {
	return int(m.count.Load())
}

func (m *memoryRequestCounter) Inc() {
	m.count.Add(1)
}

func (m *memoryRequestCounter) Desc() {
	if m.count.Load() > 0 {
		m.count.And(-1)
	}
}

func (w *WeightedLeastRequest) Next(ctx context.Context, request object.LLMRequest) string {
	if len(w.servers) == 0 {
		return ""
	}
	if len(w.servers) == 1 {
		return w.servers[0].name
	}

	selectedServer := w.servers[w.current]
	leastLoadRatio := float64(-1)

	selected := w.current

	for i, s := range w.servers {
		loadRatio := float64(s.requestCounter.Current()) / float64(s.weight)
		requestLess := loadRatio == leastLoadRatio && s.requestCounter.Less(selectedServer.requestCounter)

		if leastLoadRatio == -1 || loadRatio < leastLoadRatio || requestLess {
			selectedServer = s
			leastLoadRatio = loadRatio
			selected = i
		}
	}

	w.current = selected

	selectedServer.requestCounter.Inc()

	return selectedServer.name
}

func (w *WeightedLeastRequest) Done(_ context.Context) {
	w.servers[w.current].requestCounter.Desc()
}

var _ LoadBalancer = (*emptyLB)(nil)

type emptyLB struct{}

func (e emptyLB) Next(context.Context, object.LLMRequest) string {
	return ""
}

func (e emptyLB) Done(context.Context) {
}

func New(router *v1alpha1.Route) LoadBalancer {
	destinations := lo.Map(router.GetTargets(), func(item *v1alpha1.RouteTarget, index int) *v1alpha1.RouteDestination {
		return item.GetDestination()
	})

	switch router.GetLoadBalancePolicy() {
	case v1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_ROUND_ROBIN:
		return NewWeightedRoundRobin(destinations)
	case v1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_LEAST_REQUEST:
		return NewWeightedLeastRequest(destinations)
	case v1alpha1.LoadBalancePolicy_LOAD_BALANCE_POLICY_UNSPECIFIED:
		return &emptyLB{}
	default:
		slog.Error("unsupported load balance policy", "policy", router.GetLoadBalancePolicy())
		return emptyLB{}
	}
}
