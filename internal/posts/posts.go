package posts

import (
	"bytes"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

type Post struct {
	Title   string
	Summary string
	Tags    []string
	Date    time.Time
	Content []byte
}

type walker struct {
	posts []*Post
}

func ParsePosts() []*Post {
	w := walker{}
	filepath.WalkDir("./posts", w.walk)
	// stort by latest
	sort.Slice(w.posts, func(i, j int) bool { return w.posts[i].Date.After(w.posts[j].Date) })
	return w.posts
}

func (w *walker) walk(s string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() {
		if !strings.HasSuffix(s, ".md") {
			return nil
		}

		content, err := os.ReadFile(s)
		if err != nil {
			log.Printf("%s: could not read\n", s)
			return err
		}

		markdown := goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
			),
		)

		var buf bytes.Buffer
		context := parser.NewContext()
		if err := markdown.Convert(content, &buf, parser.WithContext(context)); err != nil {
			log.Printf("%s: invalid markdown\n", s)
			return err
		}

		parsed := buf.Bytes()
		if len(parsed) == 0 {
			log.Printf("%s: empty\n", s)
			return nil
		}

		post := Post{
			Content: parsed,
		}

		metaData := meta.Get(context)
		if value, ok := metaData["Title"].(string); ok && len(value) > 0 {
			post.Title = value
		} else {
			log.Printf("%s: title required\n", s)
			return nil
		}

		if value, ok := metaData["Summary"].(string); ok && len(value) > 0 {
			post.Summary = value
		}

		if value, ok := metaData["Tags"].([]interface{}); ok && len(value) > 0 {
			var strValue []string = make([]string, len(value))
			for i, d := range value {
				if v, ok := d.(string); ok {
					strValue[i] = v
				}
			}
			post.Tags = strValue
		}

		if value, ok := metaData["Date"].(string); ok && len(value) > 0 {
			if dt, err := time.Parse("2006/01/02", value); err == nil {
				post.Date = dt
			}
		}

		if post.Date.IsZero() {
			fi, err := os.Stat(s)
			if err != nil {
				log.Printf("%s: err stating: %s\n", s, err)
			}
			post.Date = fi.ModTime().UTC()
		}

		if post.Date.IsZero() {
			post.Date = time.Now().UTC()
		}

		log.Printf("%s: post parsed\n", s)
		w.posts = append(w.posts, &post)
	}
	return nil
}
