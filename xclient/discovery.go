// 服务发现模块

package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int // 负载均衡策略

const (
	RandomSelect     SelectMode = iota // select randomly
	RoundRobinSelect                   // select using Robbin algorithm
)

type Discovery interface {
	Refresh() error                      // refresh from remote registry
	Update(servers []string) error       // 手动更新服务列表
	Get(mode SelectMode) (string, error) // 根据负载均衡策略，选择一个服务实例
	GetAll() ([]string, error)           // 返回所有的服务实例
}

//
// MultiServersDiscovery
//  @Description: MultiServersDiscovery is a discovery for multi servers without a registry center
// user provides the server addresses explicitly instead
//
type MultiServersDiscovery struct {
	r       *rand.Rand // generate random number
	mu      sync.Mutex // protect following
	servers []string
	index   int // record the selected position for robin algorithm
}

func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

// Refresh doesn't make sense for MultiServersDiscovery, so ignore it
func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

//
// Update
//  @Description: Update the servers of discovery dynamically if needed
//  @receiver d
//  @param servers
//  @return error
//
func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

//
// Get
//  @Description:Get a server according to mode
//  @receiver d
//  @param mode
//  @return string
//  @return error
//
func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	// TODO 获取为什么要加锁
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n] // servers could be updated, so mode n to ensure safety
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

//
// GetAll
//  @Description: returns all servers in discovery
//  @receiver d
//  @return []string
//  @return error
//
func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// return a copy of servers
	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
