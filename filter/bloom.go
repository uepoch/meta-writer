package filter

import (
	"fmt"
	"hash/crc32"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/willf/bloom"
)

type Filter interface {
	Update([]byte)
	Contains([]byte) bool
	ContainsOrUpdate([]byte) bool
}

type ShardedBloomFilter struct {
	sync.Mutex
	Filters []*bloom.BloomFilter
	Timer   *time.Ticker
	stop    chan bool
}

func (f *ShardedBloomFilter) Update(key []byte) {
	zap.L().Debug("sbf.Update called", zap.ByteString("key", key))
	i := int(crc32.ChecksumIEEE(key)) % len(f.Filters)
	f.Filters[i].Add(key)
	return
}

func (f *ShardedBloomFilter) ContainsOrUpdate(key []byte) bool {
	zap.L().Debug("sbf.ContainsOrUpdate called", zap.ByteString("key", key))
	i := int(crc32.ChecksumIEEE(key)) % len(f.Filters)
	return f.Filters[i].TestAndAdd(key)
}

func (f *ShardedBloomFilter) Contains(key []byte) bool {
	zap.L().Debug("sbf.Contains called", zap.ByteString("key", key))
	i := int(crc32.ChecksumIEEE(key)) % len(f.Filters)
	return f.Filters[i].Test(key)
}

func (f *ShardedBloomFilter) startFlusher() {
	if f.Timer == nil {
		zap.L().Panic("timer must not be nil when ShardedBloomFilter.startFlusher is called")
	}
	go func() {
		i := 0
		for {
			select {
			case <-f.stop:
				zap.L().Info("stopping sbf.flusher")
				f.Timer.Stop()
				return
			case <-f.Timer.C:
				j := i % len(f.Filters)
				f.Filters[j].ClearAll()
				zap.L().Info("flushed filter", zap.Int("filter_index", j))
				i++
			}
		}
	}()
}

func NewShardedBFilter(n uint, p float64, nfb uint, flush time.Duration) (*ShardedBloomFilter, error) {
	if flush <= 0 {
		return nil, fmt.Errorf("you must provide a positive duration for flush time")
	}
	splittedN := uint(math.Ceil(float64(n) / float64(nfb)))
	if splittedN <= 1 {
		return nil, fmt.Errorf("can't create a bloomfilter with number of elements < number of filters")
	}
	zap.L().Info("new sharded bloom filter created", zap.Uint("total_n", n), zap.Uint("split_n", splittedN), zap.Float64("p", p))
	sbf := &ShardedBloomFilter{
		sync.Mutex{},
		make([]*bloom.BloomFilter, 0, nfb),
		time.NewTicker(flush),
		make(chan bool),
	}
	m, k := bloom.EstimateParameters(splittedN, p)
	for ; nfb > 0; nfb-- {
		sbf.Filters = append(sbf.Filters, bloom.New(m, k))
	}
	sbf.startFlusher()
	return sbf, nil
}
