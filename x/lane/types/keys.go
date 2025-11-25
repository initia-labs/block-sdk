package types

const (
	// ModuleName is the name of the lane module
	ModuleName = "lane"

	// StoreKey is the default store key for the lane module
	StoreKey = ModuleName

	// RouterKey is the message route for the lane module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the lane module
	QuerierRoute = ModuleName
)

const (
	prefixParams = iota + 1
)

// KeyParams is the store key for the lane module's parameters.
var KeyParams = []byte{prefixParams}
