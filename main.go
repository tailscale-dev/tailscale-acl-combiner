package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/creachadair/jtree/ast"
	"github.com/creachadair/jtree/jwcc"
)

var (
	inParentFile = flag.String("f", "", "parent file to load from")
	inChildDir   = flag.String("d", "", "directory to process files from")
	outFile      = flag.String("o", "", "file to write output to")
	verbose      = flag.Bool("v", false, "enable verbose logging")
)

func main() {
	flag.Parse()

	var parentDoc *jwcc.Object
	var err error
	if *inParentFile != "" {
		parentDoc, err = parse(*inParentFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		parentDoc = &jwcc.Object{
			Members: make([]*jwcc.Member, 0),
		}
	}

	// TODO: missing any sections?
	// TODO: anything special to do with top-level properties - https://tailscale.com/kb/1337/acl-syntax#network-policy-options ?
	aclSections := map[string]any{
		// TODO: create a type and use reflection instead?
		"acls":            new(jwcc.Array),
		"groups":          new(jwcc.Object),
		"postures":        new(jwcc.Object),
		"tagOwners":       new(jwcc.Object),
		"autoApprovers":   nil, // "autoApprovers": new(jwcc.Object), // TODO: need to merge "routes" and "exitNodes" sub-sections
		"ssh":             new(jwcc.Array),
		"nodeAttrs":       new(jwcc.Array), // TODO: need to merge anything?
		"tests":           new(jwcc.Array),
		"extraDNSRecords": new(jwcc.Array),
	}

	childDocs, err := gatherChildren(*inChildDir)
	if err != nil {
		log.Fatal(err)
	}

	err = mergeDocs(aclSections, parentDoc, childDocs)
	if err != nil {
		log.Fatal(err)
	}

	parentDoc.Sort() // TODO: make configurable via an arg?
	outputFile(parentDoc)
}

func mergeDocs(sections map[string]any, parentDoc *jwcc.Object, childDocs []*jwcc.Object) error {
	for _, child := range childDocs {
		for sectionKey, sectionObject := range sections {
			section := child.Find(sectionKey)
			if section == nil {
				continue
			}

			switch sectionType := sectionObject.(type) {
			case *jwcc.Array:
				newArr := existingOrNewArray(*parentDoc, sectionKey)
				newArr.Values = append(newArr.Values, section.Value.(*jwcc.Array).Values...)

				index := parentDoc.IndexKey(ast.TextEqual(sectionKey))
				if index != -1 {
					parentDoc.Members[index] = &jwcc.Member{Key: section.Key, Value: newArr}
				} else {
					parentDoc.Members = append(parentDoc.Members, &jwcc.Member{Key: section.Key, Value: newArr})
				}

			case *jwcc.Object:
				newObj := existingOrNewObject(*parentDoc, sectionKey)
				for _, m := range section.Value.(*jwcc.Object).Members {
					newObj.Members = append(newObj.Members, &jwcc.Member{Key: m.Key, Value: m.Value})
				}

				index := parentDoc.IndexKey(ast.TextEqual(sectionKey))
				if index != -1 {
					parentDoc.Members[index] = &jwcc.Member{Key: section.Key, Value: newObj}
				} else {
					parentDoc.Members = append(parentDoc.Members, &jwcc.Member{Key: section.Key, Value: newObj})
				}

			default:
				return fmt.Errorf("unexpected type %T for %s", sectionType, sectionKey)
			}
		}
	}
	return nil
}

func gatherChildren(path string) ([]*jwcc.Object, error) {
	children := []*jwcc.Object{}

	logVerbose(fmt.Sprintf("Walking path [%v]...\n", *inChildDir))
	err := filepath.WalkDir(
		*inChildDir,
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(path, ".json") && !strings.HasSuffix(path, ".hujson") {
				return nil
			}

			doc, err := parse(path)
			if err != nil {
				log.Fatal(err)
			}

			children = append(children, doc)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return children, nil
}

func outputFile(doc *jwcc.Object) error {
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			return err
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		err = jwcc.Format(w, doc)
		if err != nil {
			return err
		}
		w.WriteString("\n")
		w.Flush()
	} else {
		err := jwcc.Format(os.Stdout, doc)
		if err != nil {
			return err
		}
		fmt.Printf("\n")
	}
	return nil
}

func parse(path string) (*jwcc.Object, error) {
	logVerbose(fmt.Sprintf("Parsing [%v]...\n", path))

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := jwcc.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", path, err)
	}

	root, ok := doc.Value.(*jwcc.Object)
	if !ok {
		return nil, fmt.Errorf("invalid file format: document root is %T, expected object", doc.Value)
	}

	return root, nil
}

func existingOrNewArray(doc jwcc.Object, path string) *jwcc.Array { // TODO: combine with existingOrNewObject and pass in type?
	existingSection := doc.FindKey(ast.TextEqual(path))
	if existingSection == nil {
		return new(jwcc.Array)
	}
	return existingSection.Value.(*jwcc.Array)
}

func existingOrNewObject(doc jwcc.Object, path string) *jwcc.Object {
	existingSection := doc.FindKey(ast.TextEqual(path))
	if existingSection == nil {
		return new(jwcc.Object)
	}
	return existingSection.Value.(*jwcc.Object)
}

func logVerbose(message string) {
	if *verbose {
		os.Stderr.WriteString(message)
	}
}
