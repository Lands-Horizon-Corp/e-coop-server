package registry

import (
	"context"
	"fmt"
)

func (r *Registry[TData, TResponse, TRequest]) OnCreate(context context.Context, data *TData) {
	fmt.Printf("DEBUG: OnCreate called with data: %+v\n", data)

	go func() {
		<-context.Done()
		fmt.Println("DEBUG: Context done signal received in OnCreate goroutine")

		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)

			fmt.Printf("DEBUG: Dispatching topics: %+v with payload: %+v\n", topics, payload)
			r.dispatch(topics, payload)
		} else {
			fmt.Println("DEBUG: r.dispatch is nil, skipping dispatch")
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnUpdate(context context.Context, data *TData) {
	go func() {
		<-context.Done()
		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)
			r.dispatch(topics, payload)
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnDelete(context context.Context, data *TData) {
	go func() {
		<-context.Done()
		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)
			r.dispatch(topics, payload)
		}
	}()
}
