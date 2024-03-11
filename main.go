package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
)

func extractUrlParts(url string) []string {
	re := regexp.MustCompile(`/([^/]+)|:([^/]+)`)
	matches := re.FindAllStringSubmatch(url, -1)
	if len(matches) > 0 {
		var parts []string
		for _, match := range matches {
			for i := 1; i < len(match); i++ {
				if match[i] != "" {
					parts = append(parts, match[i])
				}
			}
		}
		return parts
	}
	return []string{}
}

type Route struct {
	methods  map[string]http.HandlerFunc
	children map[string]*Route
}

type Router struct {
	children map[string]*Route
}

func NewRouter() *Router {
	return &Router{make(map[string]*Route)}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	parts := extractUrlParts(req.URL.Path)
	root := router.children
	size := len(parts)
	for i, s := range parts {
		if root[s] != nil && size == i+1 {
			fmt.Println(root[s])
			handler := root[s].methods[req.Method]
			if handler == nil {
				io.WriteString(w, "NOT FOUND")
			} else {
				root[s].methods[req.Method](w, req)
			}
		} else if size == i+1 {
			io.WriteString(w, "NOT FOUND")
		}
	}
}

func (router *Router) add(path string, method string, handler http.HandlerFunc) {
	parts := extractUrlParts(path)
	root := router.children
	for _, s := range parts {
		if root[s] == nil {
			router.children[s] = &Route{
				methods: map[string]http.HandlerFunc{
					method: handler,
				},
				children: make(map[string]*Route),
			}
		} else {
			if root[s] != nil && root[s].methods[method] == nil {
				root[s].methods[method] = handler
			} else {
				root = router.children
			}
		}
	}
}

func getProducts(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "products")
}

func postProducts(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "products post")
}

func getProductsHola(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hola")
}

func main() {
	router := NewRouter()
	router.add("/products", http.MethodGet, getProducts)
	router.add("/products/hola", http.MethodGet, getProductsHola)
	router.add("/products", http.MethodPost, postProducts)
	l, err := net.Listen("tcp", ":3333")
	if err != nil {
		fmt.Printf("[ERROR] starting server: %s\n", err)
	}

	println("My first server started on", l.Addr().String())
	if err := http.Serve(l, router); err != nil {
		fmt.Printf("server closed: %s\n", err)
	}
	os.Exit(1)
}
