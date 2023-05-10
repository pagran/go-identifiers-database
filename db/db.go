package db

//go:generate go run ../data/prepare.go -i ../data/identifiers.txt -o identifiers -p db -m GetIdentifiers
//go:generate go run ../data/prepare.go -i ../data/filenames.txt -o filenames -p db -m GetFilenames
//go:generate go run ../data/prepare.go -i ../data/packages.txt -o packages -p db -m GetPackages
//go:generate go run ../data/prepare.go -i ../data/domains.txt -o domains -p db -m GetDomains
