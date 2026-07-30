package domain

type FileItem struct {
	RelativePath string
	Name         string
	Extension    string
	Size         int64
	// Deleted marks a file that lives in the Google Drive trash. It is
	// intentionally excluded from compare/resize matching keys so a trashed
	// file can still be paired with its active counterpart.
	Deleted bool
}
