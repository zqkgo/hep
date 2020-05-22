// Copyright ©2020 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rtree

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ReadVar describes a variable to be read out of a tree.
type ReadVar struct {
	Name  string      // name of the branch to read
	Leaf  string      // name of the leaf to read
	Value interface{} // pointer to the value to fill

	count string // name of the leaf-count, if any
	leaf  Leaf   // leaf to which this read-var is bound
}

// NewReadVars returns the complete set of ReadVars to read all the data
// contained in the provided Tree.
func NewReadVars(t Tree) []ReadVar {
	var vars []ReadVar
	for _, b := range t.Branches() {
		for _, leaf := range b.Leaves() {
			ptr := newValue(leaf)
			cnt := ""
			if leaf.LeafCount() != nil {
				cnt = leaf.LeafCount().Name()
			}
			vars = append(vars, ReadVar{Name: b.Name(), Leaf: leaf.Name(), Value: ptr, count: cnt})
		}
	}

	return vars
}

// Deref returns the value pointed at by this read-var.
func (rv ReadVar) Deref() interface{} {
	return reflect.ValueOf(rv.Value).Elem().Interface()
}

// ReadVarsFromStruct returns a list of ReadVars bound to the exported fields
// of the provided pointer to a struct value.
//
// ReadVarsFromStruct panicks if the provided value is not a pointer to
// a struct value.
func ReadVarsFromStruct(ptr interface{}) []ReadVar {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		panic(fmt.Errorf("rtree: expect a pointer value, got %T", ptr))
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		panic(fmt.Errorf("rtree: expect a pointer to struct value, got %T", ptr))
	}

	var (
		rt     = rv.Type()
		rvars  = make([]ReadVar, 0, rt.NumField())
		reDims = regexp.MustCompile(`\w*?\[(\w*)\]+?`)
	)

	split := func(s string) (string, []string) {
		n := s
		if i := strings.Index(s, "["); i > 0 {
			n = s[:i]
		}

		out := reDims.FindAllStringSubmatch(s, -1)
		if len(out) == 0 {
			return n, nil
		}

		dims := make([]string, len(out))
		for i := range out {
			dims[i] = out[i][1]
		}
		return n, dims
	}

	for i := 0; i < rt.NumField(); i++ {
		var (
			ft = rt.Field(i)
			fv = rv.Field(i)
		)
		if ft.Name != strings.Title(ft.Name) {
			// not exported. ignore.
			continue
		}
		rvar := ReadVar{
			Name:  nameOf(ft),
			Value: fv.Addr().Interface(),
		}

		if strings.Contains(rvar.Name, "[") {
			switch ft.Type.Kind() {
			case reflect.Slice:
				sli, dims := split(rvar.Name)
				if len(dims) > 1 {
					panic(fmt.Errorf("rtree: invalid number of slice-dimensions for field %q: %q", ft.Name, rvar.Name))
				}
				rvar.Name = sli
				rvar.count = dims[0]

			case reflect.Array:
				arr, dims := split(rvar.Name)
				if len(dims) > 3 {
					panic(fmt.Errorf("rtree: invalid number of array-dimension for field %q: %q", ft.Name, rvar.Name))
				}
				rvar.Name = arr
			default:
				panic(fmt.Errorf("rtree: invalid field type for %q, or invalid struct-tag %q: %T", ft.Name, rvar.Name, fv.Interface()))
			}
		}
		switch ft.Type.Kind() {
		case reflect.Int, reflect.Uint, reflect.UnsafePointer, reflect.Uintptr, reflect.Chan, reflect.Interface:
			panic(fmt.Errorf("rtree: invalid field type for %q: %T", ft.Name, fv.Interface()))
		case reflect.Map:
			panic(fmt.Errorf("rtree: invalid field type for %q: %T (not yet supported)", ft.Name, fv.Interface()))
		}

		rvar.Leaf = rvar.Name
		rvars = append(rvars, rvar)
	}
	return rvars
}

func nameOf(field reflect.StructField) string {
	tag, ok := field.Tag.Lookup("groot")
	if ok {
		return tag
	}
	return field.Name
}

func bindRVarsTo(t Tree, rvars []ReadVar) []ReadVar {
	ors := make([]ReadVar, 0, len(rvars))
	var flatten func(b Branch, rvar ReadVar) []ReadVar
	flatten = func(br Branch, rvar ReadVar) []ReadVar {
		nsub := len(br.Branches())
		subs := make([]ReadVar, 0, nsub)
		rv := reflect.ValueOf(rvar.Value).Elem()
		get := func(name string) int {
			rt := rv.Type()
			for i := 0; i < rt.NumField(); i++ {
				ft := rt.Field(i)
				nn := nameOf(ft)
				if nn == name {
					// exact match.
					return i
				}
				// try to remove any [xyz][range].
				// do it after exact match not to shortcut arrays
				if idx := strings.Index(nn, "["); idx > 0 {
					nn = string(nn[:idx])
				}
				if nn == name {
					return i
				}
			}
			return -1
		}

		for _, sub := range br.Branches() {
			bn := sub.Name()
			if strings.Contains(bn, ".") {
				toks := strings.Split(bn, ".")
				bn = toks[len(toks)-1]
			}
			j := get(bn)
			if j < 0 {
				continue
			}
			fv := rv.Field(j)
			bname := sub.Name()
			lname := sub.Name()
			if prefix := br.Name() + "."; strings.HasPrefix(bname, prefix) {
				bname = string(bname[len(prefix):])
			}
			if idx := strings.Index(bname, "["); idx > 0 {
				bname = string(bname[:idx])
			}
			if idx := strings.Index(lname, "["); idx > 0 {
				lname = string(lname[:idx])
			}
			leaf := sub.Leaf(lname)
			count := ""
			if leaf != nil {
				if lc := leaf.LeafCount(); lc != nil {
					count = lc.Name()
				}
			}
			subrv := ReadVar{
				Name:  rvar.Name + "." + bname,
				Leaf:  lname,
				Value: fv.Addr().Interface(),
				leaf:  leaf,
				count: count,
			}
			switch len(sub.Branches()) {
			case 0:
				subs = append(subs, subrv)
			default:
				subs = append(subs, flatten(sub, subrv)...)
			}
		}
		return subs
	}

	for i := range rvars {
		var (
			rvar = &rvars[i]
			br   = t.Branch(rvar.Name)
			leaf = br.Leaf(rvar.Leaf)
			nsub = len(br.Branches())
		)
		switch nsub {
		case 0:
			rvar.leaf = leaf
			ors = append(ors, *rvar)
		default:
			ors = append(ors, flatten(br, *rvar)...)
		}
	}
	return ors
}
