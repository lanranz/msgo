package msgo

import "strings"

//前缀树上节点
type treeNode struct {
	//节点名称，分割的路径名称
	name string
	//节点的子节点
	children   []*treeNode
	routerName string
	isEnd      bool
}

func (t *treeNode) Put(path string) {
	root := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		//分割的第一位是空格，所以跳过
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		//子节点中如果有了就跳过
		for _, node := range children {
			if node.name == name {
				isMatch = true
				//重新赋值
				t = node
				break
			}
		}
		if !isMatch {
			isEnd := false
			if index == len(strs)-1 {
				isEnd = true
			}
			node := &treeNode{name: name, children: make([]*treeNode, 0), isEnd: isEnd}
			children = append(children, node)
			t.children = children
			t = node
		}
	}
	t = root
}

func (t *treeNode) Get(path string) *treeNode {
	strs := strings.Split(path, "/")
	routerName := ""
	for index, name := range strs {
		//分割的第一位是空格，所以跳过
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		//子节点中如果有了就跳过
		for _, node := range children {
			if node.name == name ||
				node.name == "*" ||
				strings.Contains(node.name, ":") {
				isMatch = true
				routerName += "/" + node.name
				node.routerName = routerName
				//重新赋值
				t = node
				if index == len(strs)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				if node.name == "**" {
					routerName += "/" + node.name
					node.routerName = routerName
					return node
				}
			}
		}
	}
	return nil
}
