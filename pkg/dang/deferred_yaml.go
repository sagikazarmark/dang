package dang

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

func decodeYAML(data string) (any, error) {
	decoder := yaml.NewDecoder(strings.NewReader(data))

	var doc yaml.Node
	if err := decoder.Decode(&doc); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, err
	}

	var extra yaml.Node
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("trailing document")
	}

	return yamlNodeToDeferredRaw(&doc, map[*yaml.Node]bool{})
}

func encodeYAML(ctx context.Context, val Value) ([]byte, error) {
	raw, err := valueToYAMLRaw(ctx, val)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(raw)
}

func valueToYAMLRaw(ctx context.Context, val Value) (any, error) {
	switch v := val.(type) {
	case DeferredValue:
		return deferredRawToYAMLRaw(v.Raw)
	case StringValue:
		return v.Val, nil
	case EnumValue:
		return v.Val, nil
	case ScalarValue:
		return v.Val, nil
	case IntValue:
		return v.Val, nil
	case FloatValue:
		return v.Val, nil
	case BoolValue:
		return v.Val, nil
	case NullValue:
		return nil, nil
	case ListValue:
		items := make([]any, len(v.Elements))
		for i, elem := range v.Elements {
			item, err := valueToYAMLRaw(ctx, elem)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			items[i] = item
		}
		return items, nil
	case FunctionValue:
		return nil, fmt.Errorf("cannot marshal function value")
	case *ModuleValue:
		obj := make(map[string]any)
		for _, kv := range v.Bindings(PrivateVisibility) {
			if _, isFn := kv.Value.(FunctionValue); isFn {
				continue
			}
			field, err := valueToYAMLRaw(ctx, kv.Value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", kv.Key, err)
			}
			obj[kv.Key] = field
		}
		return obj, nil
	case BoundMethod:
		return nil, fmt.Errorf("cannot marshal bound method value")
	case BoundBuiltinMethod:
		return nil, fmt.Errorf("cannot marshal bound builtin method value")
	case GraphQLValue:
		var id string
		if err := v.QueryChain.Select("id").Client(v.Client).Bind(&id).Execute(ctx); err != nil {
			return nil, err
		}
		return id, nil
	default:
		return nil, fmt.Errorf("unsupported value type: %T", val)
	}
}

func deferredRawToYAMLRaw(raw any) (any, error) {
	switch v := raw.(type) {
	case json.Number:
		tag := "!!int"
		if strings.ContainsAny(v.String(), ".eE") {
			tag = "!!float"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: v.String()}, nil
	case []any:
		items := make([]any, len(v))
		for i, item := range v {
			converted, err := deferredRawToYAMLRaw(item)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			items[i] = converted
		}
		return items, nil
	case map[string]any:
		obj := make(map[string]any, len(v))
		for key, value := range v {
			converted, err := deferredRawToYAMLRaw(value)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			obj[key] = converted
		}
		return obj, nil
	default:
		return raw, nil
	}
}

func yamlNodeToDeferredRaw(node *yaml.Node, seen map[*yaml.Node]bool) (any, error) {
	if node == nil {
		return nil, nil
	}

	if seen[node] {
		return nil, fmt.Errorf("cyclic alias")
	}
	seen[node] = true
	defer delete(seen, node)

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return yamlNodeToDeferredRaw(node.Content[0], seen)

	case yaml.ScalarNode:
		switch node.Tag {
		case "!!null":
			return nil, nil
		case "!!bool":
			var b bool
			if err := node.Decode(&b); err != nil {
				return nil, err
			}
			return b, nil
		case "!!int", "!!float":
			return json.Number(normalizeYAMLNumber(node.Value)), nil
		case "!!str", "":
			return node.Value, nil
		default:
			// Keep unsupported scalar tags as strings. This keeps deferred
			// materialization in the same JSON-like domain without adding
			// YAML-specific runtime types.
			return node.Value, nil
		}

	case yaml.SequenceNode:
		items := make([]any, len(node.Content))
		for i, item := range node.Content {
			val, err := yamlNodeToDeferredRaw(item, seen)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			items[i] = val
		}
		return items, nil

	case yaml.MappingNode:
		obj := make(map[string]any, len(node.Content)/2)
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			if keyNode.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("mapping key must be a scalar")
			}
			val, err := yamlNodeToDeferredRaw(node.Content[i+1], seen)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", keyNode.Value, err)
			}
			obj[keyNode.Value] = val
		}
		return obj, nil

	case yaml.AliasNode:
		return yamlNodeToDeferredRaw(node.Alias, seen)

	default:
		return nil, fmt.Errorf("unsupported YAML node kind %d", node.Kind)
	}
}

func normalizeYAMLNumber(value string) string {
	return strings.ReplaceAll(value, "_", "")
}
