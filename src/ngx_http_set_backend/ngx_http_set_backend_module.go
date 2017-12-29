// package name: ngx_http_set_backend_module
package main

import (
	"C"
)

const (
	socket = "/Users/friparia/wrdtech.com/nginx/ngx_http_set_backend.sock"
)

//export LookupBackend
func LookupBackend(_host *C.char) *C.char {

  return C.CString(string("http://127.0.0.1:9001"))
}

func main() {
	// We need the main function to make possible
	// CGO compiler to compile the package as C shared library
}
