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

var (
	inParentFile       = flag.String("f", "", "parent file to load from")
	inChildDir         = flag.String("d", "", "directory to process files from")
	outFile            = flag.String("o", "", "file to write output to")
	verbose            = flag.Bool("v", false, "enable verbose logging")
	allowedAclSections aclSections

	// TODO: anything special to do with top-level properties - https://tailscale.com/kb/1337/acl-syntax#network-policy-options ?
	// TODO: worry about casing? mainly -allow arg not matching casing?
	preDefinedAclSections = map[string]SectionHandler{
		"acls":            handleArray(),
		"autoApprovers":   handleAutoApprovers(),
		"extraDNSRecords": handleArray(),
		"grants":          handleArray(),
		"groups":          handleObject(),
		"nodeAttrs":       handleArray(), // TODO: need to merge anything?
		"postures":        handleObject(),
		"ssh":             handleArray(),
		"tagOwners":       handleObject(),
		"tests":           handleArray(),
		"hosts":           handleObject(),
	}
)

type ParsedDocument struct {
	Path   string
	Object *jwcc.Object
}
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

func usage() {
	fmt.Fprintf(os.Stderr, "usage: tailscale-acl-combiner [flags]\n")
	flag.PrintDefaults()
}

func checkArgs() error {
	if *inParentFile == "" {
		return errors.New("missing argument -f - a parent file must be provided")
	}
	if *inChildDir == "" {
		return errors.New("missing argument -d - a directory of child files to process must be provided")
	}
	if len(allowedAclSections) == 0 {
		return errors.New("missing argument -allow - a list of acl sections to allow from children must be provided - e.g. -allow=acls,ssh")
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

	aclSections, err := getAllowedSections(allowedAclSections, preDefinedAclSections)
	if err != nil {
		log.Fatal(err)
	}

	err = mergeDocs(aclSections, parentDoc, childDocs)
	if err != nil {
		log.Fatal(err)
	}

	outputFile(parentDoc.Object)
}

func getAllowedSections(allowedAclSections []string, preDefinedAclSections map[string]SectionHandler) (map[string]SectionHandler, error) {
	aclSections := map[string]SectionHandler{}
	for _, v := range allowedAclSections {
		if preDefinedAclSections[v] == nil {
			return nil, fmt.Errorf("unsupported section [%s] specified in [-allow] flag", v)
		}
		aclSections[v] = preDefinedAclSections[v]
	}
	logVerbose("allowing ACL sections [%v]\n", aclSections)
	return aclSections, nil
}

type SectionHandler func(sectionKey string, parentPath string, parent *jwcc.Object, childPath string, childSection *jwcc.Member)

func handleArray() SectionHandler {
	return func(sectionKey string, parentPath string, parent *jwcc.Object, childPath string, childSection *jwcc.Member) {
		parentProps := parent.FindKey(ast.TextEqual(sectionKey))
		if parentProps != nil {
			pathComment(parentProps.Value.(*jwcc.Array).Values[0], parentPath)
		}

		newArr := existingOrNewArray(*parent, sectionKey)
		childArrValues := childSection.Value.(*jwcc.Array).Values

		pathCommentAlreadyAdded := false
		for i := range childArrValues {
			newArr.Values = append(newArr.Values, childArrValues[i])

			if !pathCommentAlreadyAdded {
				pathComment(childArrValues[i], childPath)
				pathCommentAlreadyAdded = true
			}

		}

		upsertMember(parent, sectionKey, newArr)
	}
}

func handleObject() SectionHandler {
	return func(sectionKey string, parentPath string, parent *jwcc.Object, childPath string, childSection *jwcc.Member) {
		parentProps := parent.FindKey(ast.TextEqual(sectionKey))
		if parentProps != nil {
			pathComment(parentProps.Value.(*jwcc.Object).Members[0], parentPath)
		}

		newObj := existingOrNewObject(*parent, sectionKey)

		pathCommentAlreadyAdded := false
		for _, m := range childSection.Value.(*jwcc.Object).Members {
			newMember := &jwcc.Member{Key: m.Key, Value: m.Value}
			newMember.Comments().Before = m.Comments().Before
			newMember.Comments().Line = m.Comments().Line
			newMember.Comments().End = m.Comments().End
			newObj.Members = append(newObj.Members, newMember)

			if !pathCommentAlreadyAdded {
				pathComment(newMember, childPath)
				pathCommentAlreadyAdded = true
			}
		}

		upsertMember(parent, sectionKey, newObj)
	}
}

func handleAutoApprovers() SectionHandler {
	// https://tailscale.com/kb/1337/acl-syntax#auto-approvers-autoapprovers
	return func(sectionKey string, parentPath string, parent *jwcc.Object, childPath string, childSection *jwcc.Member) {
		newObj := existingOrNewObject(*parent, sectionKey)

		childSectionObj := childSection.Value.(*jwcc.Object)

		childExitNodeProps := childSectionObj.FindKey(ast.TextEqual("exitNode"))
		arrayFn := handleArray()
		arrayFn("exitNode", parentPath, newObj, childPath, childExitNodeProps)

		childRoutesProps := childSectionObj.FindKey(ast.TextEqual("routes"))
		objectFn := handleObject()
		objectFn("routes", parentPath, newObj, childPath, childRoutesProps)

		newObj.Sort()
		upsertMember(parent, sectionKey, newObj)
	}
}

func upsertMember[V *jwcc.Object | *jwcc.Array](doc *jwcc.Object, key string, val V) {
	keyAst := ast.String(key)
	index := doc.IndexKey(ast.TextEqual(key))
	newMember := &jwcc.Member{Key: keyAst.Quote(), Value: jwcc.Value(val)}
	newMember.Comments().Before = jwcc.Value(val).Comments().Before
	newMember.Comments().Line = jwcc.Value(val).Comments().Line
	newMember.Comments().End = jwcc.Value(val).Comments().End
	if index != -1 {
		doc.Members[index] = newMember
	} else {
		doc.Members = append(doc.Members, newMember)
	}
}

func pathComment(val jwcc.Value, path string) {
	val.Comments().Before = append([]string{fmt.Sprintf("from `%s`", path)}, val.Comments().Before...)
}

func mergeDocs(sections map[string]SectionHandler, parentDoc *ParsedDocument, childDocs []*ParsedDocument) error {
	for _, parentSection := range parentDoc.Object.Members {
		pathComment(parentSection, parentDoc.Path)
	}
	for _, child := range childDocs {
		if child.Path == parentDoc.Path {
			logVerbose("skipping [%s], same doc as parent\n", child.Path)
			continue
		}

		for sectionKey, handlerFn := range sections {
			childSection := child.Object.Find(sectionKey)
			if childSection == nil {
				continue
			}

			handlerFn(sectionKey, parentDoc.Path, parentDoc.Object, child.Path, childSection)
			child.Object.Members = removeMember(child.Object, sectionKey)
		}

		for _, remainingSection := range child.Object.Members {
			return fmt.Errorf("unsupported section [\"%s\"] in file [%s]", remainingSection.Key, child.Path)
		}
	}

	parentDoc.Object.Sort()

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
		return nil, fmt.Errorf("invalid file format: document root is [%T], expected [object]", doc.Value)
	}

	return &ParsedDocument{Path: path, Object: root}, nil
}

func existingOrNewArray(doc jwcc.Object, key string) *jwcc.Array { // TODO: combine with existingOrNewObject and pass in type?
	existingSection := doc.FindKey(ast.TextEqual(key))
	if existingSection == nil {
		logVerbose("section [%s] not found in parent doc, creating new array\n", key)
		return new(jwcc.Array)
	}
	logVerbose("section [%s] found in parent doc, re-using array\n", key)
	return existingSection.Value.(*jwcc.Array)
}

func existingOrNewObject(doc jwcc.Object, key string) *jwcc.Object {
	existingSection := doc.FindKey(ast.TextEqual(key))
	if existingSection == nil {
		logVerbose("section [%s] not found in parent doc, creating new object\n", key)
		return new(jwcc.Object)
	}
	logVerbose("section [%s] found in parent doc, re-using object\n", key)
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
