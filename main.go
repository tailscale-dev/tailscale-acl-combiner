package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/creachadair/jtree/jwcc"
)

var (
	dev = flag.Bool("dev", false, "enable dev mode")
)

func main() {
	flag.Parse()

	// if *dev {
	// }

	// dir := "acls"
	// fmt.Printf("Parsing acl parts from [%s]\n", dir)

	files := []string{
		"acls/group1/acls.hujson",
		"acls/group2/acls.hujson",
	}

	parentDoc := &jwcc.Object{
		Members: make([]*jwcc.Member, 0),
	}

	newAcls := new(jwcc.Array)
	newGroups := new(jwcc.Object)

	for _, f := range files {
		doc, err := parse(f)
		if err != nil {
			log.Fatal(err)
		}

		acls := doc.Find("acls")
		if acls != nil {
			aclsValues := acls.Value.(*jwcc.Array)
			newAcls.Values = append(newAcls.Values, aclsValues.Values...)
		}

		groups := doc.Find("groups")
		if groups != nil {
			groupsValues := groups.Value.(*jwcc.Object)
			for _, v := range groupsValues.Members {
				newGroups.Members = append(newGroups.Members, &jwcc.Member{Key: v.Key, Value: v.Value})
			}
		}
	}

	parentDoc.Members = append(parentDoc.Members, jwcc.Field("acls", newAcls))
	parentDoc.Members = append(parentDoc.Members, jwcc.Field("groups", newGroups))

	err := jwcc.Format(os.Stdout, parentDoc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n")
}

func parse(path string) (*jwcc.Object, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := jwcc.Parse(f)
	if err != nil {
		return nil, err
	}

	root, ok := doc.Value.(*jwcc.Object)
	if !ok {
		return nil, errors.New(fmt.Sprintf("invalid policy: document root is %T, not object", doc.Value))
	}

	return root, nil
}
