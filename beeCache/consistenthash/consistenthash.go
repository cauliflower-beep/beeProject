package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32 所谓"hash"值，本质就是一串二进制的数
type Hash func(data []byte) uint32 // 采取依赖注入的方式 允许替换成自定义的Hash函数

// Map contains all hashed keys
type Map struct {
	hash     Hash
	replicas int            // 虚拟节点倍数
	keys     []int          // Sorted 哈希环
	hashMap  map[int]string // 虚拟节点与真实节点的映射表 key-虚拟节点的哈希值 value-真实节点的名称
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE // 默认为 crc32.ChecksumIEEE
	}
	return m
}

// Add 添加真实节点/机器 允许传入0或多个真实节点的名称
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { // 对每个真实节点key，对应创建 m.replicas 个虚拟节点，虚拟节点通过添加编号的方式区分不同虚拟节点
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash) // 计算虚拟节点的哈希值，并添加到环上
			m.hashMap[hash] = key         // 增加虚拟节点和真实节点的映射关系
		}
	}
	sort.Ints(m.keys) // 环上的哈希值进行排序
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key))) // 计算key的hash值
	// Search 使用二分查找的方式返回[0,n)中满足f(i)为true的最小索引i. 如果f(i) == true,那么f(i+1) == true.
	// 如果不存在这样的i，则返回n，注意不是返回 -1 之类的 Not Found
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash // 顺时针找到第一个匹配的虚拟节点的下标 idx
	})

	// 如果idx == len(m.keys) 说明应该选择m.keys[0]
	// 因为m.keys 是一个环状结构，所以用取余数的方式来处理这种情况
	return m.hashMap[m.keys[idx%len(m.keys)]] // 通过hashMap映射得到真实的节点
}
