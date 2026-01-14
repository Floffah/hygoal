package protogen

import "fmt"

func (p *Parser) Parse() (*FileNode, error) {
	file := &FileNode{}

	for !p.expect(TokenEOF) {
		node, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if node != nil {
			file.Expressions = append(file.Expressions, node)
		} else {
			if p.expect(TokenEOF) {
				break
			}
			return nil, p.getErrorf("unexpected token: %s", p.curTok.Value)
		}
	}

	return file, nil
}

func (p *Parser) parseExpression() (Node, error) {
	if p.expect(TokenKeyword) {
		if p.curTok.Value == "enum" {
			return p.parseEnum()
		} else if p.curTok.Value == "packet" {
			return p.parsePacket()
		} else if p.curTok.Value == "type" {
			return p.parseType()
		} else {
			return nil, p.getErrorf("unexpected keyword: %s", p.curTok.Value)
		}
	}

	return nil, p.getErrorf("unexpected token: %s", p.curTok.Value)
}

func (p *Parser) parseEnum() (Node, error) {
	if !p.expect(TokenKeyword) || p.curTok.Value != "enum" {
		return nil, p.getErrorf("expected 'enum' but got %s", p.curTok.Value)
	}
	p.next() // advance after confirming 'enum' keyword

	if !p.expect(TokenIdent) {
		return nil, p.getErrorf("expected enum name but got %s", p.curTok.Value)
	}
	enumName := p.curTok.Value
	p.next() // advance after reading enum name

	if !p.expect(TokenLBrace) {
		return nil, p.getErrorf("expected '{' but got %s", p.curTok.Value)
	}
	p.next() // advance after reading '{'

	enumNode := &EnumNode{
		Name:   enumName,
		Values: []EnumValueNode{},
	}

	for !p.expect(TokenRBrace) {
		if !p.expect(TokenIdent) {
			return nil, p.getErrorf("expected enum value but got %s", p.curTok.Value)
		}
		valueName := p.curTok.Value
		p.next() // advance after reading enum value

		enumNode.Values = append(enumNode.Values, EnumValueNode{Name: valueName})

		if p.expect(TokenComma) {
			p.next() // advance after reading ','
		}
	}

	if !p.expect(TokenRBrace) {
		return nil, p.getErrorf("expected '}' but got %s", p.curTok.Value)
	}
	p.next() // advance after reading '}'

	return enumNode, nil
}

func (p *Parser) parsePacket() (Node, error) {
	if !p.expect(TokenKeyword) || p.curTok.Value != "packet" {
		return nil, p.getErrorf("expected 'packet' but got %s", p.curTok.Value)
	}
	p.next() // advance after confirming 'packet' keyword

	if !p.expect(TokenNumber) {
		return nil, p.getErrorf("expected packet ID but got %s", p.curTok.Value)
	}
	packetIDStr := p.curTok.Value
	packetID, err := parseUint32(packetIDStr)
	if err != nil {
		return nil, p.getErrorf("invalid packet ID: %s", packetIDStr)
	}
	p.next() // advance after reading packet ID

	if !p.expect(TokenIdent) {
		return nil, p.getErrorf("expected packet name but got %s", p.curTok.Value)
	}
	packetName := p.curTok.Value
	p.next() // advance after reading packet name

	if !p.expect(TokenLBrace) {
		return nil, p.getErrorf("expected '{' but got %s", p.curTok.Value)
	}
	p.next() // advance after reading '{'

	packetNode := &PacketNode{
		Name:   packetName,
		ID:     uint32(packetID),
		Fields: []FieldNode{},
	}

	for !p.expect(TokenRBrace) {
		fieldNode, err := p.parseField()
		if err != nil {
			return nil, err
		}
		packetNode.Fields = append(packetNode.Fields, *fieldNode)
	}

	if !p.expect(TokenRBrace) {
		return nil, p.getErrorf("expected '}' but got %s", p.curTok.Value)
	}
	p.next() // advance after reading '}'

	return packetNode, nil
}

func (p *Parser) parseType() (Node, error) {
	if !p.expect(TokenKeyword) || p.curTok.Value != "type" {
		return nil, p.getErrorf("expected 'type' but got %s", p.curTok.Value)
	}
	p.next() // advance after confirming 'type' keyword

	if !p.expect(TokenIdent) {
		return nil, p.getErrorf("expected type name but got %s", p.curTok.Value)
	}
	typeName := p.curTok.Value
	p.next() // advance after reading type name

	if !p.expect(TokenLBrace) {
		return nil, p.getErrorf("expected '{' but got %s", p.curTok.Value)
	}
	p.next() // advance after reading '{'

	typeNode := &TypeNode{
		Name:   typeName,
		Fields: []FieldNode{},
	}

	for !p.expect(TokenRBrace) {
		fieldNode, err := p.parseField()
		if err != nil {
			return nil, err
		}
		typeNode.Fields = append(typeNode.Fields, *fieldNode)
	}

	if !p.expect(TokenRBrace) {
		return nil, p.getErrorf("expected '}' but got %s", p.curTok.Value)
	}
	p.next() // advance after reading '}'

	return typeNode, nil
}

func (p *Parser) parseField() (*FieldNode, error) {
	isFixed := true
	if p.expect(TokenAt) {
		isFixed = false
		p.next() // advance after reading '@'
	}

	if !p.expect(TokenIdent) {
		return nil, p.getErrorf("expected field name but got %s", p.curTok.Value)
	}
	fieldName := p.curTok.Value
	p.next() // advance after reading field name

	isOptional := false
	if p.expect(TokenOptional) {
		isOptional = true
		p.next() // advance after reading '?'
	}

	fieldType, err := p.parseFieldType()
	if err != nil {
		return nil, err
	}

	fieldNode := &FieldNode{
		Name:     fieldName,
		Type:     *fieldType,
		Optional: isOptional,
		Fixed:    isFixed,
	}

	return fieldNode, nil
}

func (p *Parser) parseFieldType() (*FieldTypeNode, error) {
	if !p.expect(TokenIdent) {
		return nil, p.getErrorf("expected field type but got %s", p.curTok.Value)
	}
	typeName := p.curTok.Value
	p.next() // advance after reading type name

	var minSize *int
	var maxSize *int
	if p.expect(TokenLBracket) {
		p.next() // advance after reading '['

		if !p.expect(TokenNumber) {
			return nil, p.getErrorf("expected bit size but got %s", p.curTok.Value)
		}
		lowSizeStr := p.curTok.Value
		lowSize, err := parseInt(lowSizeStr)
		if err != nil {
			return nil, p.getErrorf("invalid bit size: %s", lowSizeStr)
		}
		p.next()

		if p.expect(TokenColon) {
			p.next() // advance after reading ':'

			if !p.expect(TokenNumber) {
				return nil, p.getErrorf("expected max bit size but got %s", p.curTok.Value)
			}
			highSizeStr := p.curTok.Value
			highSize, err := parseInt(highSizeStr)
			if err != nil {
				return nil, p.getErrorf("invalid max bit size: %s", highSizeStr)
			}
			p.next()
			minSize = &lowSize
			maxSize = &highSize
		} else {
			maxSize = &lowSize
		}
		if !p.expect(TokenRBracket) {
			return nil, p.getErrorf("expected ']' but got %s", p.curTok.Value)
		}
		p.next() // advance after reading ']'
	}

	fieldTypeNode := &FieldTypeNode{
		Name:    typeName,
		MinSize: minSize,
		MaxSize: maxSize,
	}

	return fieldTypeNode, nil
}

func parseInt(s string) (int, error) {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func parseUint32(s string) (uint32, error) {
	var value uint32
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}
