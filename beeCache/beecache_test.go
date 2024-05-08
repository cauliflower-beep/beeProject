package beecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	/*
		还是标注一下吧
		下面这段代码做了两件事：
		1.使用GetterFunc()将一个匿名函数强转成了GetterFunc函数；
		2.GetterFunc函数是一个实现了Getter接口的接口型函数，所以类型转换之后，可以直接赋值给 f 这个Getter接口。
	*/
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		// 匿名函数的逻辑是，把key转换为字节切片返回
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

// db 模拟耗时的数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// TestGet 创建Group示例，测试Get方法
func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db)) // 某个键调用回调函数的次数 次数大于1 表示调用了多次回调函数 没有返回
	bee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1 // 某个key每调用一次回调，次数+1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		// 测试缓存为空的情况下，能够通过回调函数获取到源数据
		if view, err := bee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		} // load from callback function

		/**************执行到这里，缓存中已经有了3个同学的成绩***************/

		// 测试缓存已经存在的情况下，是否能够直接从缓存中获取 如果打印出了cache miss，那就没办法直接从缓存中获取
		if _, err := bee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	// 测试缓存不存在的时候，是否会返回一个错误值
	if view, err := bee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
