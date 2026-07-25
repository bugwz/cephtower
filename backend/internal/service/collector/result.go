package collector

type dataFetchResult struct {
	source          string
	recordsUpserted int
	recordsDeleted  int
}
