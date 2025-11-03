package registry

import "context"

// Find
func (r *Registry[TData, TResponse, TRequest]) Find(
	context context.Context,
	fields *TData,
	preloads ...string,
)

// FindRaw
// FindFilter
// FindFilterRaw
