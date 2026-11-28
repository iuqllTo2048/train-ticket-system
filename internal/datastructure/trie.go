package datastructure

import "sort"

// TrieNode 字典树节点。
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

// Trie 字典树。
type Trie struct {
	root *TrieNode
}

// 创建一个新节点。
func newTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
	}
}

// NewTrie 创建空字典树。
func NewTrie() *Trie {
	return &Trie{
		root: newTrieNode(),
	}
}

// Insert 插入一个完整字符串。
func (t *Trie) Insert(word string) {
	if word == "" {
		return
	}

	node := t.root

	for _, ch := range word {
		if node.children[ch] == nil {
			node.children[ch] = newTrieNode()
		}

		node = node.children[ch]
	}

	node.isEnd = true
}

// findNode 查找指定路径对应的节点。
func (t *Trie) findNode(path string) *TrieNode {
	node := t.root

	for _, ch := range path {
		next := node.children[ch]

		if next == nil {
			return nil
		}

		node = next
	}

	return node
}

// Search 判断完整字符串是否存在。
func (t *Trie) Search(word string) bool {
	node := t.findNode(word)

	return node != nil && node.isEnd
}

// SearchPrefix 查询指定前缀的完整字符串。
func (t *Trie) SearchPrefix(
	prefix string,
	limit int,
) []string {
	result := make([]string, 0)

	if limit <= 0 {
		return result
	}

	node := t.findNode(prefix)

	if node == nil {
		return result
	}

	t.collect(node, prefix, limit, &result)

	return result
}

// collect 使用 DFS 收集完整字符串。
func (t *Trie) collect(
	node *TrieNode,
	path string,
	limit int,
	result *[]string,
) {
	// 达到结果数量限制，结束递归。
	if len(*result) >= limit {
		return
	}

	// 当前路径是一个完整字符串。
	if node.isEnd {
		*result = append(*result, path)
	}

	// 收集所有子节点对应的字符。
	keys := make([]rune, 0, len(node.children))

	for ch := range node.children {
		keys = append(keys, ch)
	}

	// 排序，保证搜索结果顺序稳定。
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	// 递归遍历子节点。
	for _, ch := range keys {
		if len(*result) >= limit {
			return
		}

		t.collect(
			node.children[ch],
			path+string(ch),
			limit,
			result,
		)
	}
}
