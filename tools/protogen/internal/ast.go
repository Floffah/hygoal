package protogen

// contains the DSL ast definitions and parser logic
type Node interface {
	isNode() bool
}

type FileNode struct {
	Expressions []Node
}

func (f *FileNode) isNode() bool {
	return true
}

func (f *FileNode) FindEnum(name string) *EnumNode {
	for _, expr := range f.Expressions {
		if enum, ok := expr.(*EnumNode); ok {
			if enum.Name == name {
				return enum
			}
		}
	}
	return nil
}

func (f *FileNode) FindType(name string) *TypeNode {
	for _, expr := range f.Expressions {
		if typeN, ok := expr.(*TypeNode); ok {
			if typeN.Name == name {
				return typeN
			}
		}
	}
	return nil
}

func (f *FileNode) FindPacket(name string) *PacketNode {
	for _, expr := range f.Expressions {
		if packet, ok := expr.(*PacketNode); ok {
			if packet.Name == name {
				return packet
			}
		}
	}
	return nil
}

func (f *FileNode) FindAny(name string) Node {
	for _, expr := range f.Expressions {
		switch node := expr.(type) {
		case *EnumNode:
			if node.Name == name {
				return node
			}
		case *TypeNode:
			if node.Name == name {
				return node
			}
		case *PacketNode:
			if node.Name == name {
				return node
			}
		}
	}
	return nil
}

type EnumNode struct {
	Name   string
	Values []EnumValueNode
}

func (e *EnumNode) isNode() bool {
	return true
}

type EnumValueNode struct {
	Name string
	//Value int
}

func (e *EnumValueNode) isNode() bool {
	return true
}

type PacketNode struct {
	Name   string
	ID     uint32
	Fields []FieldNode
}

func (p *PacketNode) isNode() bool {
	return true
}

type TypeNode struct {
	Name   string
	Fields []FieldNode
}

func (t *TypeNode) isNode() bool {
	return true
}

type FieldNode struct {
	Name string
	Type FieldTypeNode
	//Repeated bool
	Optional bool
	Fixed    bool
}

func (f *FieldNode) isNode() bool {
	return true
}

type FieldTypeNode struct {
	Name    string
	MinSize *int // if min is null, size is fixed to MaxSize
	MaxSize *int
}

func (f *FieldTypeNode) isNode() bool {
	return true
}
