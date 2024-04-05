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
	postTempl "github.com/pnmcosta/csta.dev/internal/templates/post"
	tagTempl "github.com/pnmcosta/csta.dev/internal/templates/tag"
)

type Post = posts.Post

var devFlag = flag.Bool("dev", false, "if true public folder will be served")

func main() {
	flag.Parse()

	posts := posts.ParsePosts()

	// Output path.
	rootPath := "public"
	if err := os.Mkdir(rootPath, 0755); err != nil && !os.IsExist(err) {
		log.Fatalf("failed to create output directory: %v", err)
	}

	tags := map[string][]Post{}

	// Create a page for each post.
	for _, post := range posts {
		// Create the output directory.
		dir := path.Join(rootPath, post.Date.Format("2006/01/02"), slug.Make(post.Title))
		if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
			log.Fatalf("failed to create dir %q: %v", dir, err)
			continue
		}

		// Create the output file.
		name := path.Join(dir, "index.html")
		f, err := os.Create(name)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
			continue
		}

		// Create an unsafe component containing raw HTML.
		content := postTempl.Unsafe(string(post.Content))

		// Use templ to render the template containing the raw HTML.
		err = postTempl.View(post, content).Render(context.Background(), f)
		if err != nil {
			log.Fatalf("failed to write output file: %v", err)
			continue
		}

		for _, tag := range post.Tags {
			tags[tag] = append(tags[tag], post)
		}
	}

	// Create a page for each tag.
	tagKeys := make([]string, 0, len(tags))
	for tag, posts := range tags {
		slug := slug.Make(tag)

		// Create the output directory.
		dir := path.Join(rootPath, "tag", slug)
		if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
			log.Fatalf("failed to create dir %q: %v", dir, err)
			continue
		}

		// Create the output file.
		name := path.Join(dir, "index.html")
		f, err := os.Create(name)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
			continue
		}

		// Use templ to render the template containing the raw HTML.
		err = tagTempl.View(tag, posts).Render(context.Background(), f)
		if err != nil {
			log.Fatalf("failed to write output file: %v", err)
			continue
		}

		tagKeys = append(tagKeys, tag)
	}

	// Create an index page.
	name := path.Join(rootPath, "index.html")
	f, err := os.Create(name)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}

	// Write it out.
	err = index.View(posts, tagKeys).Render(context.Background(), f)
	if err != nil {
		log.Fatalf("failed to write index page: %v", err)
	}

	if !*devFlag {
		return
	}

	// Static serve only on dev
	http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Listening on 127.0.0.1:3000")
	err = http.ListenAndServe("127.0.0.1:3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
