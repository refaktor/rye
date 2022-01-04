// Go cayley example

func main() {

	tmpdir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	err = graph.InitQuadStore("bolt", tmpdir, nil)
	if err != nil {
		log.Fatal(err)
	}

	store, err := cayley.NewGraph("bolt", tmpdir, nil)
	if err != nil {
		log.Fatal(err)
	}

	store.AddQuad(quad.Make("phrase of the day", "is of course", "Hello BoltDB!", "demo graph"))

	p := cayley.StartPath(store, quad.String("phrase of the day")).Out(quad.String("is of course"))

	it, _ := p.BuildIterator().Optimize()

	defer it.Close()

	ctx := context.TODO()
	for it.Next(ctx) {
		token := it.Result()                // get a ref to a node (backend-specific)
		value := store.NameOf(token)        // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		fmt.Println(nativeValue)
	}
	if err := it.Err(); err != nil {
		log.Fatalln(err)
	}
}





