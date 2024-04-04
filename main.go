package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gosimple/slug"
	"github.com/pnmcosta/csta.dev/internal/posts"
	"github.com/pnmcosta/csta.dev/internal/templates/index"
	templates "github.com/pnmcosta/csta.dev/internal/templates/post"
)

var devFlag = flag.Bool("dev", false, "if true public folder will be served")

func main() {
	flag.Parse()

	posts := posts.ParsePosts()

	// Output path.
	rootPath := "public"
	if err := os.Mkdir(rootPath, 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("failed to create output directory: %v", err)
		}
	}

	// Create an index page.
	name := path.Join(rootPath, "index.html")
	f, err := os.Create(name)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}

	// Write it out.
	err = index.View(posts).Render(context.Background(), f)
	if err != nil {
		log.Fatalf("failed to write index page: %v", err)
	}

	// Create a page for each post.
	for _, post := range posts {
		// Create the output directory.
		dir := path.Join(rootPath, post.Date.Format("2006/01/02"), slug.Make(post.Title))
		if err := os.MkdirAll(dir, 0755); err != nil && err != os.ErrExist {
			log.Fatalf("failed to create dir %q: %v", dir, err)
		}

		// Create the output file.
		name := path.Join(dir, "index.html")
		f, err := os.Create(name)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}

		// Create an unsafe component containing raw HTML.
		content := templates.Unsafe(string(post.Content))

		// Use templ to render the template containing the raw HTML.
		err = templates.View(post, content).Render(context.Background(), f)
		if err != nil {
			log.Fatalf("failed to write output file: %v", err)
		}
	}

	if !*devFlag {
		return
	}

	http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Print("Listening on 127.0.0.1:3000...")
	err = http.ListenAndServe("127.0.0.1:3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
