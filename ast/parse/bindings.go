package main

type Binding struct {
	NodeLoc
	Name          AstName   `json:"name"`
	Annotation    *AstType  `json:"annotation"`
	ColonPosition *Position `json:"colonPosition"`
}

type BindingList []Binding
