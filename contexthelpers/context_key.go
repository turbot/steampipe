package contexthelpers

//https://medium.com/@matryer/context-keys-in-go-5312346a868d
type ContextKey string

func (c ContextKey) String() string {
	return "steampipe context key " + string(c)
}
