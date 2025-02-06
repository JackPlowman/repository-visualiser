package visualiser

type LanguageCount struct {
	Language string
	Count    int
}

type LanguageCountArray []LanguageCount

type FileStat struct {
	Path  string
	Lines int
}
