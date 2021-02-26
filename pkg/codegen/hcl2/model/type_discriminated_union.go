// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/v2/codegen/hcl2/syntax"
)

// UnionType represents values that may be any one of a specified set of types.
type DiscriminatedUnionType struct {
	Discriminator string
	Mapping map[string]*ObjectType

	s string
}

// NewUnionType creates a new union type with the given element types. Any element types that are union types are
// replaced with their element types.
func NewDiscriminatedUnionType(discriminator string, mapping map[string]*ObjectType) Type {
	return &DiscriminatedUnionType{Discriminator: discriminator, Mapping: mapping}
}

// SyntaxNode returns the syntax node for the type. This is always syntax.None.
func (*DiscriminatedUnionType) SyntaxNode() hclsyntax.Node {
	return syntax.None
}

// Traverse attempts to traverse the union type with the given traverser. This always fails.
func (t *DiscriminatedUnionType) Traverse(traverser hcl.Traverser) (Traversable, hcl.Diagnostics) {
	var types []Type
	for _, t := range t.Mapping {
		// Note that we intentionally drop errors here and assume that the traversal will dynamically succeed.
		et, diags := t.Traverse(traverser)
		if !diags.HasErrors() {
			types = append(types, et.(Type))
		}
	}

	switch len(types) {
	case 0:
		return DynamicType, hcl.Diagnostics{unsupportedReceiverType(t, traverser.SourceRange())}
	case 1:
		if types[0] == NoneType {
			return DynamicType, hcl.Diagnostics{unsupportedReceiverType(t, traverser.SourceRange())}
		}
		return types[0], nil
	default:
		return NewUnionType(types...), nil
	}
}

// Equals returns true if this type has the same identity as the given type.
func (t *DiscriminatedUnionType) Equals(other Type) bool {
	return t.equals(other, nil)
}

func (t *DiscriminatedUnionType) equals(other Type, seen map[Type]struct{}) bool {
	if t == other {
		return true
	}
	otherUnion, ok := other.(*DiscriminatedUnionType)
	if !ok {
		return false
	}
	if t.Discriminator != otherUnion.Discriminator {
		return false
	}
	if len(t.Mapping) != len(otherUnion.Mapping) {
		return false
	}
	for i, t := range t.Mapping {
		if !t.equals(otherUnion.Mapping[i], seen) {
			return false
		}
	}
	return true
}

// AssignableFrom returns true if this type is assignable from the indicated source type. A union(T_0, ..., T_N)
// from values of type union(U_0, ..., U_M) where all of U_0 through U_M are assignable to some type in
// (T_0, ..., T_N) and V where V is assignable to at least one of (T_0, ..., T_N).
func (t *DiscriminatedUnionType) AssignableFrom(src Type) bool {
	return assignableFrom(t, src, func() bool {
		for _, t := range t.Mapping {
			if t.AssignableFrom(src) {
				return true
			}
		}
		return false
	})
}

// ConversionFrom returns the kind of conversion (if any) that is possible from the source type to this type. A union
// type is convertible from a source type if any of its elements are convertible from the source type. If any element
// type is safely convertible, the conversion is safe; if no element is safely convertible but some element is unsafely
// convertible, the conversion is unsafe.
func (t *DiscriminatedUnionType) ConversionFrom(src Type) ConversionKind {
	return t.conversionFrom(src, false, nil)
}

func (t *DiscriminatedUnionType) conversionFrom(src Type, unifying bool, seen map[Type]struct{}) ConversionKind {
	return conversionFrom(t, src, unifying, seen, func() ConversionKind {
		var conversionKind ConversionKind
		for _, t := range t.Mapping {
			if ck := t.conversionFrom(src, unifying, seen); ck > conversionKind {
				conversionKind = ck
			}
		}
		return conversionKind
	})
}

// If all conversions to a dest type from a union type are safe, the conversion is safe.
// If no conversions to a dest type from a union type exist, the conversion does not exist.
// Otherwise, the conversion is unsafe.
func (t *DiscriminatedUnionType) conversionTo(dest Type, unifying bool, seen map[Type]struct{}) ConversionKind {
	conversionKind, exists := SafeConversion, false
	for _, t := range t.Mapping {
		switch dest.conversionFrom(t, unifying, seen) {
		case SafeConversion:
			exists = true
		case UnsafeConversion:
			conversionKind, exists = UnsafeConversion, true
		case NoConversion:
			conversionKind = UnsafeConversion
		}
	}
	if !exists {
		return NoConversion
	}
	return conversionKind
}

func (t *DiscriminatedUnionType) String() string {
	return t.string(nil)
}

func (t *DiscriminatedUnionType) string(seen map[Type]struct{}) string {
	if t.s == "" {
		var items []string
		for n, e := range t.Mapping {
			es := fmt.Sprintf("%s=%s", n, e.string(seen))
			items = append(items, es)
		}
		t.s = fmt.Sprintf("discriminated_union(%s, %s)", t.Discriminator, strings.Join(items, ", "))
	}
	return t.s
}

func (t *DiscriminatedUnionType) unify(other Type) (Type, ConversionKind) {
	return unify(t, other, func() (Type, ConversionKind) {
		return t.unifyTo(other)
	})
}

func (t *DiscriminatedUnionType) unifyTo(other Type) (Type, ConversionKind) {
	// Unify the other type with each element of the union and return a new union type.
	mapping, conversionKind := make(map[string]*ObjectType, len(t.Mapping)), SafeConversion
	for n, t := range t.Mapping {
		element, ck := t.unify(other)
		if ck < conversionKind {
			conversionKind = ck
		}
		mapping[n] = element.(*ObjectType)
	}
	return NewDiscriminatedUnionType(t.Discriminator, mapping), conversionKind
}

func (*DiscriminatedUnionType) isType() {}
