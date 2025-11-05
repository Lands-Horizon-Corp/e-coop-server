package registry

// ToModel converts a data entity to its response representation.
func (c *Registry[TData, TResponse, TRequest]) ToModel(data *TData) *TResponse {
	if data == nil {
		return nil
	}
	return c.resource(data)
}

// ToModels converts a slice of data entities to their response representations.
func (c *Registry[TData, TResponse, TRequest]) ToModels(data []*TData) []*TResponse {
	if data == nil {
		return []*TResponse{}
	}
	out := make([]*TResponse, 0, len(data))
	for _, item := range data {
		if m := c.ToModel(item); m != nil {
			out = append(out, m)
		}
	}
	return out
}

// ToModel converts a data entity to its response representation using a custom resource function.
func ToModel[TData any, TResponse any](data *TData, resource func(*TData) *TResponse) *TResponse {
	if data == nil {
		return nil
	}
	return resource(data)
}

// ToModels converts a slice of data entities to their response representations using a custom resource function.
func ToModels[TData any, TResponse any](data []*TData, resource func(*TData) *TResponse) []*TResponse {
	if data == nil {
		return []*TResponse{}
	}
	out := make([]*TResponse, 0, len(data))
	for _, item := range data {
		if m := ToModel(item, resource); m != nil {
			out = append(out, m)
		}
	}
	return out
}
