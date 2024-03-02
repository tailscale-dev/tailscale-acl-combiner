package main

import (
	"bufio"
	"errors"
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

const (
	typeArray  = "Array"
	typeObject = "Object"
)

var (
	inParentFile       = flag.String("f", "", "parent file to load from")
	inChildDir         = flag.String("d", "", "directory to process files from")
	outFile            = flag.String("o", "", "file to write output to")
	verbose            = flag.Bool("v", false, "enable verbose logging")
	allowedAclSections aclSections
)

type aclSections []string

func (i *aclSections) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *aclSections) Set(value string) error {
	values := strings.Split(value, ",")
	for _, v := range values {
		*i = append(*i, v)
	}
	return nil
}

type ParsedDocument struct {
	Path   string
	Object *jwcc.Object
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: tailscale-acl-combiner [flags]\n")
	flag.PrintDefaults()
}

func checkArgs() error {
	if *inChildDir == "" {
		return errors.New("missing argument -d - no directory provided to process files from")
	}
	return nil
}

func main() {
	flag.Var(&allowedAclSections, "allow", "acl sections to allow from children")
	flag.Parse()
	argsErr := checkArgs()
	if argsErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", argsErr)
		usage()
		os.Exit(1)
	}

	var parentDoc *ParsedDocument
	var err error
	if *inParentFile != "" {
		parentDoc, err = parse(*inParentFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		parentDoc = &ParsedDocument{
			Object: &jwcc.Object{
				Members: make([]*jwcc.Member, 0),
			},
		}
	}

	childDocs, err := gatherChildren(*inChildDir)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: missing any sections?
	// TODO: anything special to do with top-level properties - https://tailscale.com/kb/1337/acl-syntax#network-policy-options ?
	// TODO: worry about casing? mainly -allow arg not maching casing?
	preDefinedAclSections := map[string]string{
		"acls": typeArray,
		// "autoApprovers" - autoApprovers should not be delegated (until we get feedback that they should)
		"extraDNSRecords": typeArray,
		"grants":          typeArray,
		"groups":          typeObject,
		"nodeAttrs":       typeArray, // TODO: need to merge anything?
		"postures":        typeObject,
		"ssh":             typeArray,
		"tagOwners":       typeObject,
		"tests":           typeArray,
	}

	aclSections := getAllowedSections(allowedAclSections, preDefinedAclSections)
	err = mergeDocs(aclSections, parentDoc, childDocs)
	if err != nil {
		log.Fatal(err)
	}

	parentDoc.Object.Sort()
	outputFile(parentDoc.Object)
}

func getAllowedSections(allowedAclSections []string, preDefinedAclSections map[string]string) map[string]string {
	aclSections := map[string]string{}
	// TODO: handle `newsection:Array` as input?
	for _, v := range allowedAclSections {
		aclSections[v] = preDefinedAclSections[v]
	}
	logVerbose("allowing ACL sections [%v]\n", aclSections)
	return aclSections
}

func mergeDocs(sections map[string]string, parentDoc *ParsedDocument, childDocs []*ParsedDocument) error {
	for _, child := range childDocs {
		// TODO: insert 'from <file>' comment into new doc
		for sectionKey, sectionObject := range sections {
			section := child.Object.Find(sectionKey)
			if section == nil {
				continue
			}

			if sectionObject == typeArray {
				newArr := existingOrNewArray(*parentDoc.Object, sectionKey)
				newArr.Values = append(newArr.Values, section.Value.(*jwcc.Array).Values...)

				index := parentDoc.Object.IndexKey(ast.TextEqual(sectionKey))
				if index != -1 {
					parentDoc.Object.Members[index] = &jwcc.Member{Key: section.Key, Value: newArr}
				} else {
					parentDoc.Object.Members = append(parentDoc.Object.Members, &jwcc.Member{Key: section.Key, Value: newArr})
				}
			} else if sectionObject == typeObject {
				newObj := existingOrNewObject(*parentDoc.Object, sectionKey)
				for _, m := range section.Value.(*jwcc.Object).Members {
					newObj.Members = append(newObj.Members, &jwcc.Member{Key: m.Key, Value: m.Value})
				}

				index := parentDoc.Object.IndexKey(ast.TextEqual(sectionKey))
				if index != -1 {
					parentDoc.Object.Members[index] = &jwcc.Member{Key: section.Key, Value: newObj}
				} else {
					parentDoc.Object.Members = append(parentDoc.Object.Members, &jwcc.Member{Key: section.Key, Value: newObj})
				}
			} else {
				return fmt.Errorf("unexpected type [%v] for [\"%s\"] from file [%s]", sectionObject, sectionKey, parentDoc.Path)
			}

			child.Object.Members = removeMember(child.Object, sectionKey)
		}

		for _, remainingSection := range child.Object.Members {
			// TODO: arg to log and not error on unsupported sections?
			return fmt.Errorf("unsupported section [\"%s\"] in file [%s]", remainingSection.Key, child.Path)
		}
	}
	return nil
}

func gatherChildren(path string) ([]*ParsedDocument, error) {
	children := []*ParsedDocument{}

	logVerbose(fmt.Sprintf("walking path [%v]...\n", path))
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

func parse(path string) (*ParsedDocument, error) {
	logVerbose(fmt.Sprintf("parsing [%v]...\n", path))

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

	return &ParsedDocument{Path: path, Object: root}, nil
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

func removeMember(obj *jwcc.Object, key string) []*jwcc.Member {
	indexKey := obj.IndexKey(ast.TextEqual(key))

	if indexKey == -1 {
		return obj.Members
	}

	ret := make([]*jwcc.Member, 0)
	ret = append(ret, obj.Members[:indexKey]...)
	return append(ret, obj.Members[indexKey+1:]...)
}

func logVerbose(message string, a ...any) {
	if *verbose {
		os.Stderr.WriteString(fmt.Sprintf(message, a...))
	}
}
