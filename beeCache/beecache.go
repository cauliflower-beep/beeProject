/*负责与外部交互，控制缓存存储和获取的主流程*/

package beecache

import (
	"fmt"
	"log"
	"sync"
)

// Getter loads data for a key
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error) // 函数类型实现了Getter接口，称GetterFunc为接口型函数

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter // 缓存未命中时获取源数据的回调（callback）
	mainCache cache
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter") // 要求必须定义缓存未命中时的回调函数
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,                          // 每个group拥有一个唯一的name
		getter:    getter,                        // 缓存未命中时获取源数据的回调
		mainCache: cache{cacheBytes: cacheBytes}, // 并发缓存
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock() // 不涉及任何冲突变量的写操作，故用只读锁 RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 先从本节点读缓存
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[BeeCache] hit")
		return v, nil
	}

	// 没有的话再尝试读取远程节点的缓存
	return g.load(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
// 将HTTPPool注入到Group中 HTTPPool实现了接口PeerPicker
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// load 使用PickPeer()方法选择节点
func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			// 若非本机节点，则调用getFromPeer()从远程获取
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[BeeCache] Failed to get from peer", err)
		}
	}
	// 若是本机节点或远程获取失败，回退到getLocally
	return g.getLocally(key)
}

// getFromPeer 使用 httpGetter 访问远程节点，获取缓存值 httpGetter 实现了 PeerGetter 接口
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
