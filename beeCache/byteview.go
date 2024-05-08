/*缓存值的抽象与封装*/

package beecache

// ByteView A ByteView holds an immutable view of bytes.
type ByteView struct {
	b []byte // 存储真实的缓存值 选byte类型是为了能够支持存储字符串、图片等任意的数据类型的存储
}

// Len returns the view's length
func (v ByteView) Len() int {
	return len(v.b) // remember? day1中我们规定，缓存对象必须实现Value接口
}

// ByteSlice returns a copy of the data as a byte slice.
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b) // b是只读的，所以使用ByteSlice返回一个copy，防止缓存值被外部程序修改
}

// String returns the data as a string, making a copy if necessary.
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
