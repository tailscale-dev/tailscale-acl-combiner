package main

import (
	"fmt"
	"os"
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
		t.Fatalf("expected no error, got %v", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	sections := map[string]SectionHandler{
		"goodpath": handleObject(),
	}

	err = mergeDocs(sections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(parentDoc.Object.Members) != 1 {
		t.Fatalf("parent members length should be 1, got %v", len(parentDoc.Object.Members))
	}

	if parentDoc.Object.IndexKey(ast.TextEqual("goodpath")) != 0 {
		t.Fatalf("section index key length should be 0, got %v", parentDoc.Object.IndexKey(ast.TextEqual("goodpath")))
	}
}

func TestMergeDocsParentWithDifferentMembers(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"otherpath": {"foo":"bar", "bar":"foo"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	sections := map[string]SectionHandler{
		"goodpath": handleObject(),
	}

	err = mergeDocs(sections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(parentDoc.Object.Members) != 2 {
		t.Fatalf("parent members length should be 2, got %v", len(parentDoc.Object.Members))
	}
}

func TestMergeDocsParentWithSameMember(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"bar":"foo"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	sections := map[string]SectionHandler{
		"goodpath": handleObject(),
	}

	err = mergeDocs(sections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(parentDoc.Object.Members) != 1 {
		t.Fatalf("parent members length should be 1, got %v", len(parentDoc.Object.Members))
	}

	memberIndexKey := parentDoc.Object.IndexKey(ast.TextEqual("goodpath"))
	if memberIndexKey != 0 {
		t.Fatalf("section index key length should be 0, got %v", memberIndexKey)
	}

	member := parentDoc.Object.Members[memberIndexKey]
	memberObjectMembers := member.Value.(*jwcc.Object).Members
	if len(memberObjectMembers) != 2 {
		t.Fatalf("member object keys length should be 2, got %v", len(memberObjectMembers))
	}
}

func TestExistingOrNewObject(t *testing.T) {
	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
	}

	goodpathObject := existingOrNewObject(*childDoc.Object, "goodpath")
	if len(goodpathObject.Members) != 1 {
		t.Fatalf("object members length should be 1, got %v", len(goodpathObject.Members))
	}

	badpathObject := existingOrNewObject(*childDoc.Object, "badpath")
	if len(badpathObject.Members) != 0 {
		t.Fatalf("object members length should be 0, got %v", len(badpathObject.Members))
	}
}

func TestExistingOrNewArray(t *testing.T) {
	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": ["bar"]
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
	}

	goodpathObject := existingOrNewArray(*childDoc.Object, "goodpath")
	if len(goodpathObject.Values) != 1 {
		t.Fatalf("object members length should be 1, got %v", len(goodpathObject.Values))
	}

	badpathObject := existingOrNewArray(*childDoc.Object, "badpath")
	if len(badpathObject.Values) != 0 {
		t.Fatalf("object members length should be 0, got %v", len(badpathObject.Values))
	}
}

func TestRemoveMember(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"bar":"foo"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
	}

	sameMembers := removeMember(parentDoc.Object, "NOTHING_TO_REMOVE")
	if len(sameMembers) != 1 {
		t.Fatalf("members count should be [%v], got %v", 1, len(sameMembers))
	}

	removedMembers := removeMember(parentDoc.Object, "goodpath")
	if len(removedMembers) != 0 {
		t.Fatalf("members count should be [%v], got %v", 0, len(removedMembers))
	}
}

func TestGetAllowedSections(t *testing.T) {
	actualValue := handleObject()
	defined := map[string]SectionHandler{
		"1": actualValue,
		"2": actualValue,
		"3": actualValue,
	}
	allowed := []string{"1", "2"}
	allowedAclSections := getAllowedSections(allowed, defined)

	// should exist
	section1 := allowedAclSections["1"]
	if section1 == nil {
		t.Fatalf("section [%v] should NOT be nil", "1")
	}
	section2 := allowedAclSections["2"]
	if section2 == nil {
		t.Fatalf("section [%v] should NOT be nil", "2")
	}

	// should not exist
	section3 := allowedAclSections["3"]
	if section3 != nil {
		t.Fatalf("section [%v] SHOULD be nil", "3")
	}
	sectionZ := allowedAclSections["Z"]
	if sectionZ != nil {
		t.Fatalf("section [%v] SHOULD be nil", "Z")
	}
}

func TestArrayHandler(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		// empty parent
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"acls": [
			{"action": "accept", "src": ["finance1"], "dst": ["tag:demo-infra:22"]},
		]
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	childSection := child.Value.(*jwcc.Object).Find("acls")

	handlerFn := handleArray()
	handlerFn("acls", parentDoc.Path, parentDoc.Object, "CHILD", childSection)

	aclMembers := parentDoc.Object.Find("acls").Value.(*jwcc.Array).Values

	os.Stderr.WriteString(fmt.Sprintf("aclMembers: %v", aclMembers))

	if len(aclMembers) != 1 {
		t.Fatalf("section [%v] should be 1, not [%v]", "acls", len(aclMembers))
	}
}
