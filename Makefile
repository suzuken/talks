deps:
	which present || go get -u golang.org/x/tools/cmd/present
	which misspell || go get -u github.com/client9/misspell/cmd/misspell

present:
	present

check:
	misspell */*.slide
