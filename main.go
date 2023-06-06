package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/spec", specHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func specHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "err: %v", err)
		return
	}

	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	f, err := url.Parse(r.FormValue("uri"))
	if err != nil {
		fmt.Fprintf(w, "err: %v", err)
		return
	}
	doc, err := loader.LoadFromURI(f)
	if err != nil {
		fmt.Fprintf(w, "err: %v", err)
		return
	}

	// Validate document
	err = doc.Validate(ctx)
	if err != nil {
		fmt.Fprintf(w, "err: %v", err)
		return
	}

	fmt.Fprintf(w, "title: %s\n", doc.Info.Title)
	fmt.Fprintf(w, "description: %s\n", doc.Info.Description)
	fmt.Fprintf(w, "version: %s\n", doc.Info.Version)
	fmt.Fprint(w, "\n")

	for path, info := range doc.Paths {
		for method, operation := range info.Operations() {
			fmt.Fprintf(w, "path: %s\n", path)
			fmt.Fprintf(w, "method: %s\n", method)
			fmt.Fprintf(w, "description: %s\n", strings.Split(operation.Description, "\n")[0])
			if len(operation.Parameters) > 0 {
				fmt.Fprint(w, "params:\n")
				for _, param := range operation.Parameters {
					fmt.Fprintf(w, " - %s | %s | %s | %s\n", param.Value.Name, param.Value.In, param.Value.Schema.Value.Type, param.Value.Description)
					//spew.Dump(param.Value)
				}
			}
			if operation.RequestBody != nil {
				fmt.Fprint(w, "body: ")
				for _, contentType := range operation.RequestBody.Value.Content {
					fmt.Fprintf(w, "%s\n", contentType.Schema.Value.Type)
					if len(contentType.Schema.Value.Properties) > 0 {
						for name, prop := range contentType.Schema.Value.Properties {
							fmt.Fprintf(w, " - %s: %s\n", name, prop.Value.Type)
						}
					}
					//spew.Dump(contentType.Value.Schema.Value)
				}
			}
			if len(operation.Responses) > 0 {
				fmt.Fprint(w, "responses: \n")
				for code, response := range operation.Responses {
					fmt.Fprintf(w, " - %s (%s)\n", code, *response.Value.Description)
					if len(response.Value.Content) > 0 {
						for name, contentType := range response.Value.Content {
							fmt.Fprintf(w, "   - %s: %s\n", name, contentType.Schema.Value.Type)
						}
					}
					//spew.Dump(response.Value)
				}
			}
			fmt.Fprint(w, "\n")
		}
		//spew.Dump(info)
	}
	//fmt.Fprintf(w, "schema : %v", doc.Paths)
	return
}
