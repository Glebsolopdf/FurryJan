package api

type Post struct {
	ID    int   `json:"id"`
	Tags  Tags  `json:"tags"`
	File  File  `json:"file"`
	Flags Flags `json:"flags"`
	Score Score `json:"score"`
}

type Score struct {
	Up    int `json:"up"`
	Down  int `json:"down"`
	Total int `json:"total"`
}

type Tags struct {
	General   []string `json:"general"`
	Species   []string `json:"species"`
	Character []string `json:"character"`
	Copyright []string `json:"copyright"`
	Artist    []string `json:"artist"`
	Lore      []string `json:"lore"`
	Meta      []string `json:"meta"`
}

type File struct {
	URL  string `json:"url"`
	Ext  string `json:"ext"`
	Size int    `json:"size"`
	MD5  string `json:"md5"`
}

type Flags struct {
	Pending      bool `json:"pending"`
	Flagged      bool `json:"flagged"`
	NoteLocked   bool `json:"note_locked"`
	StatusLocked bool `json:"status_locked"`
}

type PostsResponse struct {
	Posts []Post `json:"posts"`
}

type PostResponse struct {
	Post Post `json:"post"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}
