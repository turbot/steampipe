package db_common

// Query is struct to encapsulate a query, with any required params
type Query struct {
	QueryString string
	Args        []any
}
