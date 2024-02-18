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

	f1 := "acls/group1/acls.hujson"
	d1, err := parse(f1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("loaded doc [%v] from [%s]\n", d1, f1)
	a1 := d1.Find("acls").Value.(*jwcc.Array)
	fmt.Printf("loaded acls [%v] from [%s]\n", a1, f1)

	f2 := "acls/group2/acls.hujson"
	d2, err := parse(f2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("loaded doc [%v] from [%s]\n", d2, f2)
	a2 := d2.Find("acls").Value.(*jwcc.Array)
	fmt.Printf("loaded acls [%v] from [%s]\n", a2, f2)

	newDoc := &jwcc.Object{
		Members: make([]*jwcc.Member, 0),
	}

	root := new(jwcc.Array)
	root.Values = append(root.Values, a1.Values...)
	root.Values = append(root.Values, a2.Values...)

	newDoc.Members = append(newDoc.Members, jwcc.Field("acls", root))

	// fmt.Printf("new acls %+v\n", root.JSON())

	err = jwcc.Format(os.Stdout, newDoc)
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
