package datastructure

import (
	"reflect"
	"testing"
)

// 测试字符串插入与精确查找。
func TestTrieSearch(t *testing.T) {
	trie := NewTrie()

	trie.Insert("成都东")
	trie.Insert("成都南")

	if !trie.Search("成都东") {
		t.Fatal("成都东应该存在")
	}

	if !trie.Search("成都南") {
		t.Fatal("成都南应该存在")
	}

	if trie.Search("成都") {
		t.Fatal("成都目前只是前缀，不是完整字符串")
	}

	if trie.Search("重庆北") {
		t.Fatal("重庆北不应该存在")
	}

	// 将原来的前缀变成完整字符串。
	trie.Insert("成都")

	if !trie.Search("成都") {
		t.Fatal("成都插入后应该存在")
	}

	t.Log("Trie 插入与精确查找测试通过")
}

// 测试中文前缀搜索。
func TestTrieSearchPrefix(t *testing.T) {
	trie := NewTrie()

	stations := []string{
		"成都南",
		"重庆北",
		"成都市域站",
		"成都东",
		"北京南",
		"成都",
		"成都东",
	}

	for _, station := range stations {
		trie.Insert(station)
	}

	result := trie.SearchPrefix("成都", 10)

	expected := []string{
		"成都",
		"成都东",
		"成都南",
		"成都市域站",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf(
			"前缀搜索错误: expected=%v, actual=%v",
			expected,
			result,
		)
	}

	t.Logf("成都前缀搜索结果: %v", result)

	result = trie.SearchPrefix("成都市", 10)

	expected = []string{"成都市域站"}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf(
			"多层递归搜索错误: expected=%v, actual=%v",
			expected,
			result,
		)
	}

	t.Logf("成都市前缀搜索结果: %v", result)
}

// 测试数量限制与边界情况。
func TestTrieLimit(t *testing.T) {
	trie := NewTrie()

	trie.Insert("成都")
	trie.Insert("成都东")
	trie.Insert("成都南")
	trie.Insert("成都市域站")
	trie.Insert("重庆北")

	result := trie.SearchPrefix("成都", 2)

	expected := []string{
		"成都",
		"成都东",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("limit 限制错误: %v", result)
	}

	t.Logf("limit=2 搜索结果: %v", result)

	// 不存在的前缀。
	if len(trie.SearchPrefix("上海", 10)) != 0 {
		t.Fatal("不存在的前缀应该返回空结果")
	}

	// 不合法的数量限制。
	if len(trie.SearchPrefix("成都", 0)) != 0 {
		t.Fatal("limit=0 应该返回空结果")
	}

	// 空前缀代表搜索整棵树。
	result = trie.SearchPrefix("", 10)

	if len(result) != 5 {
		t.Fatalf("空前缀应该返回全部5个站点: %v", result)
	}

	t.Logf("空前缀搜索结果: %v", result)

	// 空字符串不允许插入。
	trie.Insert("")

	if trie.Search("") {
		t.Fatal("空字符串不应该存在")
	}

	t.Log("Trie 边界情况测试通过")
}
