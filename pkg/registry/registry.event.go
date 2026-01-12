package registry

import (
	"context"
	"log"
	"reflect"
)

func (r *Registry[TData, TResponse, TRequest]) OnCreate(ctx context.Context, data *TData) {
	go func() {
		<-ctx.Done()
		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)
			if err := r.dispatch(topics, payload); err != nil {
				log.Fatalf("OnCreate dispatch failed: %v - %s", err, reflect.TypeOf(*data))
			}
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnUpdate(ctx context.Context, data *TData) {
	go func() {
		<-ctx.Done()
		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)
			if err := r.dispatch(topics, payload); err != nil {
				log.Fatalf("OnUpdate dispatch failed: %v - %s", err, reflect.TypeOf(*data))
			}
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnDelete(ctx context.Context, data *TData) {
	go func() {
		<-ctx.Done()
		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)
			if err := r.dispatch(topics, payload); err != nil {
				log.Fatalf("OnDelete dispatch failed: %v - %s", err, reflect.TypeOf(*data))
			}
		}
	}()
}
