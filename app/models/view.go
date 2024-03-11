package models

type Page struct {
	Type          string
	FormType      string
	Title         string
	View          interface{}
	ErrorMessages []string
	OkMessage     string
	RequestURL    string
	Theme         string
}

// ListRowField struct
type ListRowField struct {
	Type       string
	Value      string
	Class      string
	Link       string
	Confirm    string
	Icon       string
	Modal      string
	ModalTitle string
	FieldClass string
}

// ListRow struct
type ListRow struct {
	Fields []ListRowField
}

// ListPagination struct
type ListPagination struct {
	Current  int
	Previous int
	Next     int
	Pages    []int
}

// List struct
type List struct {
	Pagination ListPagination
	Rows       []ListRow
}

// FormField struct
type FormField struct {
	Name        string
	Title       string
	Type        string
	Class       string
	Placeholder string
	Value       string
	Values      []string
	Checked     bool
	Required    bool
}

// Form struct
type Form struct {
	Fields     []FormField
	CancelLink string
	SubmitName string
}
