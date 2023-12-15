package model

type Blogs struct {
	Tags  []BlogTag
	Blogs []Blog
}

type BlogTag struct {
	Title       string
	Description string
	Items       []string
}

type Blog struct {
	Selector    []string
	Title       string
	Description string
	Date        string
	Path        string
}
