package main

//Loadable - interface for hydration
type Loadable interface {
	LoadValue(name string, value []string)
}

//Hydrate - hydrate model
func Hydrate(loadable Loadable, data map[string][]string) {
	for key, value := range data {
		loadable.LoadValue(key, value)
	}
}
