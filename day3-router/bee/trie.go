package bee

import (
	"fmt"
	"strings"
)

/*
	前缀树(Trie树)是最常用来实现动态路由的一种数据结构
	每一个节点的所有子节点都拥有相同的前缀
*/

type node struct {
	pattern  string  // 待匹配路由，例如 /p/:lang 可以匹配/p/c/doc 和 /p/go/doc
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc,tutorial,intro]
	/*
		与普通的树不同，新加的isWild字段是实现动态路由匹配的关键
		即当我们匹配 /p/go/doc/这个路由时，第一层节点，p精准匹配到了p，第二层节点，go模糊匹配到:lang，那么将会把lang这个参数赋值为go，继续下一层匹配。
	*/
	isWild bool // 是否精确匹配，part含有 : 或 * 时为true
}

// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

/*
	对于路由来说，最重要的就是注册与匹配了
	开发服务时，注册路由规则，映射handler；
	访问服务时，匹配路由规则，查找到对应的handler
	因此，Trie树需要支持节点的插入与查询。
*/
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height)
		if result != nil {
			return result
		}
	}
	return nil
}

func (n *node) travel(list *[]*node) {
	if n.pattern != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}

func (n *node) String() string {
	return fmt.Sprintf("node{pattern=%s, part=%s, isWild=%t}", n.pattern, n.part, n.isWild)
}
