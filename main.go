package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/gosimple/slug"
	"github.com/pnmcosta/csta.dev/internal/posts"
	"github.com/pnmcosta/csta.dev/internal/templates/index"
	postTempl "github.com/pnmcosta/csta.dev/internal/templates/post"
	tagTempl "github.com/pnmcosta/csta.dev/internal/templates/tag"
)

type Post = posts.Post

var devFlag = flag.Bool("dev", false, "if true public folder will be served")

func main() {
	start := time.Now().UTC()
	flag.Parse()

	posts := posts.ParsePosts()

	// Output path.
	rootPath := "public"
	if err := os.Mkdir(rootPath, 0755); err != nil && os.IsNotExist(err) {
		log.Fatalf("failed to create %q: %v", rootPath, err)
	}

	tags := map[string][]*Post{}

	var wg sync.WaitGroup
	wg.Add(len(posts))
	// Create a page for each post.
	for _, post := range posts {
		go func() {
			defer wg.Done()

			// Create the output directory.
			dir := path.Join(rootPath, post.Date.Format("2006/01/02"), slug.Make(post.Title))
			if err := os.MkdirAll(dir, 0755); err != nil && os.IsNotExist(err) {
				log.Printf("failed to create %q: %v\n", dir, err)
				return
			}

			// Create the output file.
			name := path.Join(dir, "index.html")
			f, err := os.Create(name)
			if err != nil {
				log.Printf("failed to create %q: %v\n", name, err)
				return
			}

			// Create an unsafe component containing raw HTML.
			content := postTempl.Unsafe(string(post.Content))

			// Use templ to render the template containing the raw HTML.
			err = postTempl.View(post, content).Render(context.Background(), f)
			if err != nil {
				log.Printf("failed to write %q: %v\n", name, err)
				return
			}

			log.Printf("%s created", name)
		}()

		// note failure to render will still generate tag ref
		for _, tag := range post.Tags {
			tags[tag] = append(tags[tag], post)
		}
	}
	wg.Wait()

	wg.Add(len(tags))
	// Create a page for each tag.
	tagKeys := make([]string, 0, len(tags))
	for tag, posts := range tags {
		tagKeys = append(tagKeys, tag)

		go func() {
			defer wg.Done()

			slug := slug.Make(tag)

			// Create the output directory.
			dir := path.Join(rootPath, "tag", slug)
			if err := os.MkdirAll(dir, 0755); err != nil && os.IsNotExist(err) {
				log.Printf("failed to create %q: %v\n", dir, err)
				return
			}

			// Create the output file.
			name := path.Join(dir, "index.html")
			f, err := os.Create(name)
			if err != nil {
				log.Printf("failed to create %q: %v\n", name, err)
				return
			}

			// Use templ to render the template containing the raw HTML.
			err = tagTempl.View(tag, posts).Render(context.Background(), f)
			if err != nil {
				log.Printf("failed to write %q: %v\n", name, err)
				return
			}

			log.Printf("%s created", name)
		}()
	}

	// sort tag keys by post count descending
	sort.Slice(tagKeys, func(i, j int) bool { return len(tags[tagKeys[i]]) > len(tags[tagKeys[j]]) })

	wg.Wait()

	// Create an index page.
	wg.Add(1)
	go func() {
		defer wg.Done()
		name := path.Join(rootPath, "index.html")
		f, err := os.Create(name)
		if err != nil {
			log.Printf("failed to create %q: %v\n", name, err)
			return
		}

		// Write it out.
		err = index.View(posts, tagKeys).Render(context.Background(), f)
		if err != nil {
			log.Printf("failed to write %q: %v\n", name, err)
			return
		}

		log.Printf("%s created", name)
	}()

	// Create a posts page.
	wg.Add(1)
	go func() {
		defer wg.Done()

		if len(posts) == 0 {
			return
		}

		name := path.Join(rootPath, "posts.html")
		f, err := os.Create(name)
		if err != nil {
			log.Printf("failed to create %q: %v\n", name, err)
			return
		}

		// Write it out.
		err = index.Posts(posts).Render(context.Background(), f)
		if err != nil {
			log.Printf("failed to write %q: %v\n", name, err)
			return
		}

		log.Printf("%s created", name)
	}()

	wg.Wait()

	log.Printf("took %dms\n", time.Now().UTC().Sub(start).Milliseconds())

	if !*devFlag {
		return
	}

	// Static serve only on dev
	http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Listening on http://127.0.0.1:3000")
	if err := http.ListenAndServe("127.0.0.1:3000", nil); err != nil {
		log.Fatal(err)
	}
}
