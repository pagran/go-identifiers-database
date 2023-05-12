package db

//go:generate go run ../data/prepare.go -p db -c -i ../data/identifiers.txt -o identifiers -m GetIdentifiers
//go:generate go run ../data/prepare.go -p db -c -i ../data/filenames.txt -o filenames -m GetFilenames
//go:generate go run ../data/prepare.go -p db -i ../data/packages.txt -o packages -m GetPackages
//go:generate go run ../data/prepare.go -p db -i ../data/domains.txt -o domains -m GetDomains
