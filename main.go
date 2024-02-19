package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/creachadair/jtree/jwcc"
)

var (
	f   = flag.String("f", "", "parent file to load from")
	dir = flag.String("d", "", "directory to process files from")
	// TODO: add -out arg to write to file
	verbose = flag.Bool("v", false, "enable verbose logging")
)

func main() {
	flag.Parse()

	// if *dev {
	// }
	var parentDoc *jwcc.Object
	var err error
	if *f != "" {
		parentDoc, err = parse(*f)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		parentDoc = &jwcc.Object{
			Members: make([]*jwcc.Member, 0),
		}
	}

	// TODO: add additional sections
	newAcls := new(jwcc.Array)
	newGroups := new(jwcc.Object)

	logVerbose(fmt.Sprintf("Walking path [%v]...\n", *dir))
	err = filepath.WalkDir(
		*dir,
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			logVerbose(fmt.Sprintf("Parsing [%v]...\n", path))

			doc, err := parse(path)
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

			// TODO: parse after each file to report errors found when they happen?

			return nil
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	parentDoc.Members = append(parentDoc.Members, jwcc.Field("acls", newAcls))
	parentDoc.Members = append(parentDoc.Members, jwcc.Field("groups", newGroups))

	err = jwcc.Format(os.Stdout, parentDoc)
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

func logVerbose(message string) {
	if *verbose {
		os.Stderr.WriteString(message)
	}
}
