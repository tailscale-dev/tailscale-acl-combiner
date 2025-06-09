.PHONY: testdata
testdata:
	go run . \
		-v \
		-f testdata/input-parent.hujson \
		-d testdata/departments/ \
		-allow=acls,autoApprovers,grants,groups,ipsets,ssh,tests,sshTests \
		-o testdata/output-file-to-compare-to.hujson
