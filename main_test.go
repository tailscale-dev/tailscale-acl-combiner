package main

import (
	"os"
	"strings"
	"testing"

	"github.com/creachadair/jtree/ast"
	"github.com/creachadair/jtree/jwcc"
)

const (
	ACL_PARENT = `{
	"groups": {
		"group:engineering": [
			"dave@example.com",
			"laura@example.com",
		],
		"group:sales": [
			"brad@example.com",
			"alice@example.com",
		],
	},
	"acls": [
		{
			"action": "accept",
			"src": ["group:security-team@example.com"],
			"dst": ["tag:logging:*"]
		}
	],
	"tagOwners": {
		"tag:logging": ["group:security-team@example.com"]
	},
	"autoApprovers": {
		"routes": {
			"192.0.2.0/24": ["group:engineering", "alice@example.com", "tag:foo"],
		},
		"exitNode": ["tag:bar"],
	},
}`
)

func TestMergeDocsEmptyParent(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		// empty parent
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
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
		t.Fatalf("expected no error, got [%v]", err)
	}

	if len(parentDoc.Object.Members) != 1 {
		t.Fatalf("parent members length should be [1], got [%v]", len(parentDoc.Object.Members))
	}

	if parentDoc.Object.IndexKey(ast.TextEqual("goodpath")) != 0 {
		t.Fatalf("section index key length should be [0], got [%v]", parentDoc.Object.IndexKey(ast.TextEqual("goodpath")))
	}
}

func TestMergeDocsParentWithDifferentMembers(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(ACL_PARENT))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
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
		t.Fatalf("expected no error, got [%v]", err)
	}

	if len(parentDoc.Object.Members) != 5 {
		t.Fatalf("parent members length should be [5], got [%v]", len(parentDoc.Object.Members))
	}
}

func TestMergeDocsParentWithSameMember(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"bar":"foo"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
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
		t.Fatalf("expected no error, got [%v]", err)
	}

	if len(parentDoc.Object.Members) != 1 {
		t.Fatalf("parent members length should be [1], got [%v]", len(parentDoc.Object.Members))
	}

	memberIndexKey := parentDoc.Object.IndexKey(ast.TextEqual("goodpath"))
	if memberIndexKey != 0 {
		t.Fatalf("section index key length should be [0], got [%v]", memberIndexKey)
	}

	member := parentDoc.Object.Members[memberIndexKey]
	memberObjectMembers := member.Value.(*jwcc.Object).Members
	if len(memberObjectMembers) != 2 {
		t.Fatalf("member object keys length should be [2], got [%v]", len(memberObjectMembers))
	}
}

func TestPathCommentsForObject(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"bar":"foo"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
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
		t.Fatalf("expected no error, got [%v]", err)
	}

	if len(parentDoc.Object.Members) != 1 {
		t.Fatalf("parent members length should be [1], got [%v]", len(parentDoc.Object.Members))
	}

	memberIndexKey := parentDoc.Object.IndexKey(ast.TextEqual("goodpath"))
	if memberIndexKey != 0 {
		t.Fatalf("section index key length should be [0], got [%v]", memberIndexKey)
	}

	member := parentDoc.Object.Members[memberIndexKey]
	memberObjectMembers := member.Value.(*jwcc.Object).Members
	if len(memberObjectMembers) != 2 {
		t.Fatalf("member object keys length should be [2], got [%v]", len(memberObjectMembers))
	}

	barMember := member.Value.(*jwcc.Object).Find("bar")
	if barMember.Value.String() != "foo" {
		t.Fatalf("member value should be [foo], got [%v]", barMember.Value.String())
	}
	barMemberComments := barMember.Comments().Before
	if barMemberComments[0] != "from `parent`" {
		t.Fatalf("member comment should be [from `parent`], got [%v]", barMemberComments[0])
	}

	fooMember := member.Value.(*jwcc.Object).Find("foo")
	if fooMember.Value.String() != "bar" {
		t.Fatalf("member value should be [bar], got [%v]", fooMember.Value.String())
	}
	fooMemberComments := fooMember.Comments().Before
	if fooMemberComments[0] != "from `child`" {
		t.Fatalf("member comment should be [from `child`], got [%v]", fooMemberComments[0])
	}
}

func TestPathCommentsForArray(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"things": [{"thing1":"foo"}],
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"things": [{"thing2":"bar"}],
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	sections := map[string]SectionHandler{
		"things": handleArray(),
	}

	err = mergeDocs(sections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	if len(parentDoc.Object.Members) != 1 {
		t.Fatalf("parent members length should be [1], got [%v]", len(parentDoc.Object.Members))
	}

	thingsMember := parentDoc.Object.Find("things")
	if thingsMember == nil {
		t.Fatalf("section index key length should be not nil, got [%v]", thingsMember)
	}

	thingsMemberValues := thingsMember.Value.(*jwcc.Array).Values
	if len(thingsMemberValues) != 2 {
		t.Fatalf("members length should be [2], got [%v]", len(thingsMemberValues))
	}

	barMember := thingsMemberValues[0].(*jwcc.Object)
	if barMember.Members[0].Key.String() != "thing1" {
		t.Fatalf("member key should be [thing1], got [%v]", barMember.Members[0].Key.String())
	}
	if barMember.Members[0].Value.String() != "foo" {
		t.Fatalf("member value should be [foo], got [%v]", barMember.Members[0].Value.String())
	}
	if barMember.Comments().Before[0] != "from `parent`" {
		t.Fatalf("member comment should be [from `parent`], got [%v]", barMember.Comments().Before[0])
	}

	fooMember := thingsMemberValues[1].(*jwcc.Object)
	if fooMember.Members[0].Key.String() != "thing2" {
		t.Fatalf("member key should be [thing2], got [%v]", fooMember.Members[0].Key.String())
	}
	if fooMember.Members[0].Value.String() != "bar" {
		t.Fatalf("member value should be [bar], got [%v]", fooMember.Members[0].Value.String())
	}
	if fooMember.Comments().Before[0] != "from `child`" {
		t.Fatalf("member comment should be [from `parent`], got [%v]", fooMember.Comments().Before[0])
	}
}

func TestExistingOrNewObject(t *testing.T) {
	child, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"foo":"bar"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
	}

	goodpathObject := existingOrNewObject(*childDoc.Object, "goodpath")
	if len(goodpathObject.Members) != 1 {
		t.Fatalf("object members length should be [1], got [%v]", len(goodpathObject.Members))
	}

	badpathObject := existingOrNewObject(*childDoc.Object, "badpath")
	if len(badpathObject.Members) != 0 {
		t.Fatalf("object members length should be [0], got [%v]", len(badpathObject.Members))
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
		t.Fatalf("object members length should be [1], got [%v]", len(goodpathObject.Values))
	}

	badpathObject := existingOrNewArray(*childDoc.Object, "badpath")
	if len(badpathObject.Values) != 0 {
		t.Fatalf("object members length should be [0], got [%v]", len(badpathObject.Values))
	}
}

func TestRemoveMember(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"goodpath": {"bar":"foo"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
	}

	sameMembers := removeMember(parentDoc.Object, "NOTHING_TO_REMOVE")
	if len(sameMembers) != 1 {
		t.Fatalf("members count should be [%v], got [%v]", 1, len(sameMembers))
	}

	removedMembers := removeMember(parentDoc.Object, "goodpath")
	if len(removedMembers) != 0 {
		t.Fatalf("members count should be [%v], got [%v]", 0, len(removedMembers))
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
	allowedAclSections, err := getAllowedSections(allowed, defined)
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	// should exist
	section1 := allowedAclSections["1"]
	if section1 == nil {
		t.Fatalf("section [%v] should NOT be [nil]", "1")
	}
	section2 := allowedAclSections["2"]
	if section2 == nil {
		t.Fatalf("section [%v] should NOT be [nil]", "2")
	}

	// should not exist
	section3 := allowedAclSections["3"]
	if section3 != nil {
		t.Fatalf("section [%v] SHOULD be [nil]", "3")
	}
	sectionZ := allowedAclSections["Z"]
	if sectionZ != nil {
		t.Fatalf("section [%v] SHOULD be [nil]", "Z")
	}
}

func TestGetAllowedSectionsInvalidSection(t *testing.T) {
	actualValue := handleObject()
	defined := map[string]SectionHandler{
		"1": actualValue,
		"2": actualValue,
		"3": actualValue,
	}
	allowed := []string{"1", "2", "invalid"}
	_, err := getAllowedSections(allowed, defined)
	if err == nil {
		t.Fatalf("expected error, got [%v]", err)
	}
}

func TestHandleArray(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(ACL_PARENT))
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

	mergedValues := parentDoc.Object.Find("acls").Value.(*jwcc.Array).Values
	if len(mergedValues) != 2 {
		t.Fatalf("section [%v] should be [1], not [%v]", "acls", len(mergedValues))
	}
}
func TestHandleObject(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(ACL_PARENT))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"groups": {
			"group:from_child": [
				"dave@example.com",
				"laura@example.com",
			],
		}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	childSection := child.Value.(*jwcc.Object).Find("groups")

	handlerFn := handleObject()
	handlerFn("groups", parentDoc.Path, parentDoc.Object, "CHILD", childSection)

	mergedValues := parentDoc.Object.Find("groups").Value.(*jwcc.Object).Members
	if len(mergedValues) != 3 {
		t.Fatalf("section [%v] should be [1], not [%v]", "groups", len(mergedValues))
	}
}

func TestHandleAutoApprovers(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(ACL_PARENT))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"autoApprovers": {
			"routes": {
				"10.0.1.0/24": ["group:engineering", "alice@example.com", "tag:foo"],
			},
			"exitNode": ["tag:foo"],
		},
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	childSection := child.Value.(*jwcc.Object).Find("autoApprovers")

	handlerFn := handleAutoApprovers()
	handlerFn("autoApprovers", parentDoc.Path, parentDoc.Object, "CHILD", childSection)

	mergedValues := parentDoc.Object.Find("autoApprovers").Value
	if len(mergedValues.(*jwcc.Object).Members) != 2 {
		t.Fatalf("section [%v] should be [2], not [%v]", "autoApprovers", len(mergedValues.(*jwcc.Object).Members))
	}

	routesValues := mergedValues.(*jwcc.Object).Find("routes").Value.(*jwcc.Object).Members
	if len(routesValues) != 2 {
		t.Fatalf("section [%v] should be [2], not [%v]", "routes", len(routesValues))
	}

	exitNodeValues := mergedValues.(*jwcc.Object).Find("exitNode").Value.(*jwcc.Array).Values
	if len(exitNodeValues) != 2 {
		t.Fatalf("section [%v] should be [2], not [%v]", "exitNode", len(exitNodeValues))
	}
}

func TestEmptyParentObject(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{"hosts":{}}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"hosts": {
			"host1": "100.99.98.97",
		}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	err = mergeDocs(preDefinedAclSections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	mergedValues := parentDoc.Object.Find("hosts").Value.(*jwcc.Object).Members
	if len(mergedValues) != 1 {
		t.Fatalf("section [%v] should be [1], not [%v]", "hosts", len(mergedValues))
	}
}

func TestEmptyParentArray(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{"acls":[]}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
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
		t.Fatalf("expected no error, got [%v]", err)
	}

	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	err = mergeDocs(preDefinedAclSections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	mergedValues := parentDoc.Object.Find("acls").Value.(*jwcc.Array).Values
	if len(mergedValues) != 1 {
		t.Fatalf("section [%v] should be [1], not [%v]", "acls", len(mergedValues))
	}
}

func TestEmptyChildObject(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"hosts": {
			"host1": "100.99.98.97",
		}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{"hosts":{}}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	err = mergeDocs(preDefinedAclSections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	mergedValues := parentDoc.Object.Find("hosts").Value.(*jwcc.Object).Members
	if len(mergedValues) != 1 {
		t.Fatalf("section [%v] should be [1], not [%v]", "hosts", len(mergedValues))
	}
}

func TestEmptyChildArray(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(`{
		"acls": [
			{"action": "accept", "src": ["finance1"], "dst": ["tag:demo-infra:22"]},
		]
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{"acls":[]}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	err = mergeDocs(preDefinedAclSections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	mergedValues := parentDoc.Object.Find("acls").Value.(*jwcc.Array).Values
	if len(mergedValues) != 1 {
		t.Fatalf("section [%v] should be [1], not [%v]", "acls", len(mergedValues))
	}
}

func TestSort(t *testing.T) {
	parent, err := jwcc.Parse(strings.NewReader(ACL_PARENT))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}
	parentDoc := &ParsedDocument{
		Object: parent.Value.(*jwcc.Object),
		Path:   "parent",
	}

	child, err := jwcc.Parse(strings.NewReader(`{
		"autoApprovers": {
			"routes": {
				"10.0.1.0/24": ["group:engineering", "alice@example.com", "tag:foo"],
			},
			"exitNode": ["tag:foo"],
		},
	}`))
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	childDoc := &ParsedDocument{
		Object: child.Value.(*jwcc.Object),
		Path:   "child",
	}

	err = mergeDocs(preDefinedAclSections, parentDoc, []*ParsedDocument{childDoc})
	if err != nil {
		t.Fatalf("expected no error, got [%v]", err)
	}

	expectedSort := []string{"acls", "autoApprovers", "groups", "tagOwners"}
	for i, v := range expectedSort {
		if parentDoc.Object.Members[i].Key.String() != v {
			t.Fatalf("section [%v] should be position [%v]", v, i)
		}
	}
}

func printDocument(doc *ParsedDocument) {
	err := jwcc.Format(os.Stdout, doc.Object)
	if err != nil {
		panic(err)
	}
}
