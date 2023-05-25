package db

type NameType int

const (
	Unknown NameType = iota
	File
	Package
	Func
	Type
	Param
	Var
	Field
)

//go:generate go run ../data/prepare.go -p db -e NameType -i ../data/dataset.csv -o dataset -m GetNames
