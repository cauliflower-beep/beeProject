package beecache

// PeerPicker is the interface that must be implemented to locate the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 根据传入的key选择相应节点PeerGetter
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(group string, key string) ([]byte, error) // 从对应group查找缓存值。PeerGetter就对应于缓存获取流程中的HTTP客户端
}
