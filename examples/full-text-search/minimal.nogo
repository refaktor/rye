// Go example

import "github.com/blevesearch/bleve/v2"

func main() {

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("example.bleve", mapping)

	err = index.Index("elephant and a cat", "A mouse scared an elephant. A cat caught an mouse.")
	err = index.Index("mouse and dog", "Cat was hunting for a mouse. A dog chased it away.")
	err = index.Index("elephant and a dog", "Elephant looked at the dog. The dog looked at the cat.")
	
	query := bleve.NewMatchQuery("dog")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	fmt.Println(searchResults)
}
