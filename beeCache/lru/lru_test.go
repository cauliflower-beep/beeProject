package lru

import (
	"fmt"
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("dragonBall", String("Goku"))
	if v, ok := lru.Get("dragonBall"); !ok || string(v.(String)) != "Goku" {
		fmt.Println("cache hit dragonBall = Goku failed.")
	} else {
		fmt.Println("cache hit dragonBall = Goku succ.")
	}
	if _, ok := lru.Get("OnePiece"); !ok {
		fmt.Println("cache miss OnePiece succ.")
	}
}

// TestRemoveoldest 使用内存超过了设定值，触发“无用”节点移除
func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	capicity := len(k1 + k2 + v1 + v2)
	lru := New(int64(capicity), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("k1"); !ok && lru.Len() == 2 {
		fmt.Println("removeOldest k1 succ.")
	}
}

// TestOnEvicted 回调函数能否被调用
func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, val Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if reflect.DeepEqual(expect, keys) {
		fmt.Printf("Call OnEvicted succ, expect keys equals to %s", expect)
	}
}
