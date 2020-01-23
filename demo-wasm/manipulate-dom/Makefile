all: static/main.wasm static/wasm_exec.js
ifeq (, $(shell which goexec))	
	go get -u github.com/shurcooL/goexec
endif
	goexec 'http.ListenAndServe(`:9999`, http.FileServer(http.Dir(`.`)))'

static/wasm_exec.js:
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" static

static/main.wasm: main.go
	GO111MODULE=auto GOOS=js GOARCH=wasm go build -o static/main.wasm .