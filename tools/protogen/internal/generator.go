package protogen

import (
	"bytes"
	"embed"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

//go:embed templates/*
var embedFS embed.FS

var decodeTemplate *template.Template

var stringsTemplate *template.Template
var enumTemplate *template.Template
var uuidTemplate *template.Template
var arrayTemplate *template.Template
var byteArrayTemplate *template.Template
var callTypeTemplate *template.Template

func loadTemplate(name string) *template.Template {
	templateCode, err := embedFS.ReadFile("templates/" + name + ".gotmpl")
	if err != nil {
		panic(err)
	}

	funcMap := template.FuncMap{
		"capitalize": capitalize,
		"add": func(a, b int) int {
			return a + b
		},
		"dromedary": func(in string) string {
			if len(in) == 0 {
				return in
			}
			return string(in[0]+32) + in[1:]
		},
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(templateCode))
	if err != nil {
		panic(err)
	}

	return tmpl
}

func init() {
	decodeTemplate = loadTemplate("decode_fn")

	stringsTemplate = loadTemplate("strings")
	enumTemplate = loadTemplate("enum")
	uuidTemplate = loadTemplate("uuid")
	arrayTemplate = loadTemplate("array")
	byteArrayTemplate = loadTemplate("byte_array")
	callTypeTemplate = loadTemplate("call_decode_type")
}

func GenerateGoCode(ast *FileNode) (string, error) {
	str := ""

	for _, expr := range ast.Expressions {
		switch node := expr.(type) {
		case *EnumNode:
			enumCode, err := generateEnumCode(node)
			if err != nil {
				return "", err
			}
			str += enumCode
		case *PacketNode:
			packetCode, err := generatePacketCode(ast, node)
			if err != nil {
				return "", err
			}
			str += packetCode
		case *TypeNode:
			typeCode, err := generateTypeCode(node)
			if err != nil {
				return "", err
			}
			str += typeCode
		}
	}

	return str, nil
}

func generateEnumCode(enum *EnumNode) (string, error) {
	code := "type " + enum.Name + " byte\n\n"

	code += "const (\n"
	for _, value := range enum.Values {
		code += "\t" + value.Name + " " + enum.Name + " = iota\n"
	}
	code += ")\n"

	return code, nil
}

func generateTypeCode(typeN *TypeNode) (string, error) {
	code := "type " + typeN.Name + " struct {\n"
	for _, field := range typeN.Fields {
		goType := mapFieldTypeToGoType(field.Type)
		fieldName := capitalize(field.Name)
		code += "\t" + fieldName + " " + goType + "\n"
	}
	code += "}\n"

	return code, nil
}

type DecodeData struct {
	Packet           *PacketNode
	ParsingBody      string
	SizeOfFixedFrame int
}

type FieldData struct {
	Field  *FieldNode
	Offset int
}

func generatePacketCode(file *FileNode, packet *PacketNode) (string, error) {
	code := "type " + packet.Name + " struct {\n"
	for _, field := range packet.Fields {
		goType := mapFieldTypeToGoType(field.Type)

		if field.Optional {
			goType = "*" + goType
		}

		fieldName := capitalize(field.Name)
		code += "\t" + fieldName + " " + goType + "\n"
	}
	code += "}\n\n"

	//code += "func Decode" + packet.Name + "(payload []byte) (Packet, error) {\n"
	//code += "\t// TODO: implement decoding logic\n"
	//code += "\treturn &" + packet.Name + "{}, nil\n"
	//code += "}\n\n"

	parsingBodyBuf := bytes.NewBufferString("")
	currentOffset := 1

	parsingBodyBuf.WriteString("// fixed fields\n")

	for _, field := range packet.Fields {
		if field.Fixed {
			fieldParserCode, newOffset, err := writeFieldParser(file, packet, &field, currentOffset)
			if err != nil {
				return "", err
			}
			parsingBodyBuf.WriteString(fieldParserCode)
			currentOffset = newOffset
		}
	}

	offsetCode, offset, err := writeFieldOffsets(packet, currentOffset)
	if err != nil {
		return "", err
	}
	parsingBodyBuf.WriteString("\n// offsets\n")
	parsingBodyBuf.WriteString(offsetCode)
	byteSizeOfFixedFrame := offset
	currentOffset = offset

	parsingBodyBuf.WriteString("\n// variable-length fields\n")

	nonFixedFieldIndex := 0
	for _, field := range packet.Fields {
		if !field.Fixed {
			if field.Optional {
				parsingBodyBuf.WriteString("\tif (nullBits & 0x" + fmt.Sprintf("%02X", 1<<nonFixedFieldIndex) + ") != 0 {\n")
			}

			fieldParserCode, _, err := writeFieldParser(file, packet, &field, currentOffset)
			if err != nil {
				return "", err
			}
			parsingBodyBuf.WriteString(fieldParserCode)

			if field.Optional {
				nonFixedFieldIndex++
				parsingBodyBuf.WriteString("\t}\n")
			}

			parsingBodyBuf.WriteString("\n")
		}
	}

	decodeBuf := bytes.NewBufferString("")

	templateData := DecodeData{
		Packet:           packet,
		ParsingBody:      parsingBodyBuf.String(),
		SizeOfFixedFrame: byteSizeOfFixedFrame,
	}

	err = decodeTemplate.Execute(decodeBuf, templateData)
	if err != nil {
		return "", err
	}

	code += decodeBuf.String() + "\n"

	code += "func (p *" + packet.Name + ") ID() uint32 {\n"
	code += "\treturn " + fmt.Sprintf("%d", packet.ID) + "\n"
	code += "}\n"

	return code, nil
}

func writeFieldParser(file *FileNode, _ *PacketNode, field *FieldNode, offset int) (string, int, error) {
	buf := bytes.NewBufferString("\n// Field " + field.Name + "\n")

	if field.Type.Name == "ascii" || field.Type.Name == "utf8" || field.Type.Name == "string" {
		if field.Type.MaxSize == nil {
			return "", 0, fmt.Errorf("string field %s must have a max size", field.Name)
		}

		fieldData := FieldData{Field: field, Offset: offset}
		err := stringsTemplate.Execute(buf, fieldData)
		if err != nil {
			return "", 0, err
		}

		return buf.String(), offset + *field.Type.MaxSize, nil
	} else if field.Type.Name == "uuid" {
		fieldData := FieldData{Field: field, Offset: offset}
		err := uuidTemplate.Execute(buf, fieldData)
		if err != nil {
			return "", 0, err
		}

		return buf.String(), offset + 16, nil
	} else if strings.HasPrefix(field.Type.Name, "array") {
		fieldData := FieldData{Field: field, Offset: offset}
		err := arrayTemplate.Execute(buf, fieldData)
		if err != nil {
			return "", 0, err
		}

		if field.Type.Name == "array.byte" {
			err := byteArrayTemplate.Execute(buf, fieldData)
			if err != nil {
				return "", 0, err
			}
		}

		return buf.String(), offset + *field.Type.MaxSize, nil
	}

	anyExpression := file.FindAny(field.Type.Name)

	if anyExpression != nil {
		if _, ok := anyExpression.(*EnumNode); ok {
			fieldData := FieldData{Field: field, Offset: offset}
			err := enumTemplate.Execute(buf, fieldData)
			if err != nil {
				return "", 0, err
			}

			return buf.String(), offset + 1, nil
		} else if _, ok := anyExpression.(*TypeNode); ok {
			fieldData := FieldData{Field: field, Offset: offset}
			err := callTypeTemplate.Execute(buf, fieldData)
			if err != nil {
				return "", 0, err
			}

			return buf.String(), offset, nil
		}
	}

	return "", 0, nil
}

func writeFieldOffsets(packet *PacketNode, offset int) (string, int, error) {
	buf := bytes.NewBufferString("")

	for _, field := range packet.Fields {
		if !field.Fixed {
			buf.WriteString("\t" + field.Name + "Offset := int(int32(binary.LittleEndian.Uint32(payload[" + strconv.Itoa(offset) + ":" + strconv.Itoa(offset+4) + "])))\n")
			offset += 4
		}
	}

	return buf.String(), offset, nil
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'A' && s[0] <= 'Z' {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func mapFieldTypeToGoType(fieldType FieldTypeNode) string {
	switch fieldType.Name {
	case "uint16":
		return "uint16"
	case "ascii", "utf8", "string":
		return "string"
	case "uuid":
		return "uuid.UUID"
	case "array.byte":
		return "[]byte"
	default:
		return fieldType.Name
	}
}
