package main

import "reflect"

// CstKey identifies a node in the CST map. It is the node's location record, which is a pointer unique to that node, and the node's type. Both are comparable, so the key works for a node that holds slices, which cannot be a map key itself.
type CstKey struct {
	Loc *NodeLoc
	Kind reflect.Type
}

// cstKeyOf gives the key of a node in the CST map.
func cstKeyOf(node AstNode) CstKey {
	key := CstKey{Kind: reflect.TypeOf(node)}
	value := reflect.ValueOf(node)

	if value.Kind() != reflect.Struct {
		return key
	}

	if field := value.FieldByName("NodeLoc"); field.IsValid() && !field.IsNil() {
		key.Loc = field.Interface().(*NodeLoc)
	}

	return key
}

// Cst gives the CST of a node, and whether the node has one. The CST is only stored when Options.StoreCstData is set.
func (r Result) Cst(node AstNode) (CstNode, bool) {
	cst, ok := r.CstNodeMap[cstKeyOf(node)]
	return cst, ok
}
