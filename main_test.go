package main

import (
	"strings"
	"testing"

	"github.com/creachadair/jtree/ast"
	"github.com/creachadair/jtree/jwcc"
)

func TestMergeDocsEmptyParent(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		// empty parent
	}`))
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
	}

	sections := map[string]any{
		"goodpath": new(jwcc.Object),
	}

	err = mergeDocs(sections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}

	if len(parentDoc.Object.Members) != 1 {
		t.Fatalf(`parent members length should be 1, got %v`, len(parentDoc.Object.Members))
	}

	if parentDoc.Object.IndexKey(ast.TextEqual("goodpath")) != 0 {
		t.Fatalf(`section index key length should be 0, got %v`, parentDoc.Object.IndexKey(ast.TextEqual("goodpath")))
	}
}

func TestMergeDocsParentWithDifferentMembers(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"otherpath": {"foo":"bar", "bar":"foo"}
	}`))
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
	}

	sections := map[string]any{
		"goodpath": new(jwcc.Object),
	}

	err = mergeDocs(sections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}

	if len(parentDoc.Object.Members) != 2 {
		t.Fatalf(`parent members length should be 2, got %v`, len(parentDoc.Object.Members))
	}
}

func TestMergeDocsParentWithSameMember(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"bar":"foo"}
	}`))
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
	}

	sections := map[string]any{
		"goodpath": new(jwcc.Object),
	}

	err = mergeDocs(sections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}

	if len(parentDoc.Object.Members) != 1 {
		t.Fatalf(`parent members length should be 1, got %v`, len(parentDoc.Object.Members))
	}

	memberIndexKey := parentDoc.Object.IndexKey(ast.TextEqual("goodpath"))
	if memberIndexKey != 0 {
		t.Fatalf(`section index key length should be 0, got %v`, memberIndexKey)
	}

	member := parentDoc.Object.Members[memberIndexKey]
	memberObjectMembers := member.Value.(*jwcc.Object).Members
	if len(memberObjectMembers) != 2 {
		t.Fatalf(`member object keys length should be 2, got %v`, len(memberObjectMembers))
	}
}

func TestExistingOrNewObject(t *testing.T) {
	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
	}

	goodpathObject := existingOrNewObject(*childDoc.Object, "goodpath")
	if len(goodpathObject.Members) != 1 {
		t.Fatalf(`object members length should be 1, got %v`, len(goodpathObject.Members))
	}

	badpathObject := existingOrNewObject(*childDoc.Object, "badpath")
	if len(badpathObject.Members) != 0 {
		t.Fatalf(`object members length should be 0, got %v`, len(badpathObject.Members))
	}
}

func TestExistingOrNewArray(t *testing.T) {
	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": ["bar"]
	}`))
	if err != nil {
		t.Fatalf(`expected no error, got %v`, err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
	}

	goodpathObject := existingOrNewArray(*childDoc.Object, "goodpath")
	if len(goodpathObject.Values) != 1 {
		t.Fatalf(`object members length should be 1, got %v`, len(goodpathObject.Values))
	}

	badpathObject := existingOrNewArray(*childDoc.Object, "badpath")
	if len(badpathObject.Values) != 0 {
		t.Fatalf(`object members length should be 0, got %v`, len(badpathObject.Values))
	}
}
