package ard

import (
	"gopkg.in/yaml.v3"
)

//
// Locator
//

type Locator interface {
	Locate(path ...PathElement) (int, int, bool)
}

//
// YAMLLocator
//

type YAMLLocator struct {
	RootNode *yaml.Node
}

func NewYAMLLocator(rootNode *yaml.Node) *YAMLLocator {
	return &YAMLLocator{rootNode}
}

// ([Locator] interface)
func (self *YAMLLocator) Locate(path ...PathElement) (int, int, bool) {
	if node := FindYAMLNode(self.RootNode, path...); node != nil {
		return node.Line, node.Column, true
	} else {
		return -1, -1, false
	}
}
