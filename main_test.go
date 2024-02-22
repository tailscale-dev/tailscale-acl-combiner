package main

import (
	"testing"

	"github.com/creachadair/jtree/ast"
	"github.com/creachadair/jtree/jwcc"
)

var (
	goodpath = "goodpath"
	badpath  = "badpath"
)

func TestExistingOrNewObject(t *testing.T) {
	memberObject := new(jwcc.Object)
	memberObject.Members = append(memberObject.Members, &jwcc.Member{Key: ast.String("foo"), Value: jwcc.ToValue("bar")})

	doc := &jwcc.Object{}
	doc.Members = append(doc.Members, &jwcc.Member{Key: ast.String(goodpath), Value: memberObject})

	goodpathObject := existingOrNewObject(*doc, goodpath)
	if len(goodpathObject.Members) != 1 {
		t.Fatalf(`object members should be 1, got %v`, len(goodpathObject.Members))
	}

	badpathObject := existingOrNewObject(*doc, badpath)
	if len(badpathObject.Members) != 0 {
		t.Fatalf(`object members should be 0, got %v`, len(badpathObject.Members))
	}
}

func TestExistingOrNewArray(t *testing.T) {
	memberArray := new(jwcc.Array)
	memberArray.Values = append(memberArray.Values, jwcc.ToValue("bar"))

	doc := &jwcc.Object{}
	doc.Members = append(doc.Members, &jwcc.Member{Key: ast.String(goodpath), Value: memberArray})

	goodpathObject := existingOrNewArray(*doc, goodpath)
	if len(goodpathObject.Values) != 1 {
		t.Fatalf(`object members should be 1, got %v`, len(goodpathObject.Values))
	}

	badpathObject := existingOrNewArray(*doc, badpath)
	if len(badpathObject.Values) != 0 {
		t.Fatalf(`object members should be 0, got %v`, len(badpathObject.Values))
	}
}
