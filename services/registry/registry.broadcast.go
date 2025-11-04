package registry

import (
	"context"

	"go.uber.org/zap"
)

func (r *Registry[TData, TResponse, TRequest]) OnCreate(context context.Context, entity *TData) {
	go func() {
		<-context.Done()
		topics := r.created(entity)
		payload := r.ToModel(entity)
		if err := r.service.Broker.Dispatch(context, topics, payload); err != nil {
			if r.service.Logger != nil {
				r.service.Logger.Error("OnCreate dispatch error", zap.Error(err))
			}
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnUpdate(context context.Context, entity *TData) {
	go func() {
		<-context.Done()
		topics := r.updated(entity)
		payload := r.ToModel(entity)
		if err := r.service.Broker.Dispatch(context, topics, payload); err != nil {
			if r.service.Logger != nil {
				r.service.Logger.Error("OnUpdate dispatch error", zap.Error(err))
			}
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnDelete(context context.Context, entity *TData) {
	go func() {
		<-context.Done()
		topics := r.deleted(entity)
		payload := r.ToModel(entity)
		if err := r.service.Broker.Dispatch(context, topics, payload); err != nil {
			if r.service.Logger != nil {
				r.service.Logger.Error("OnDelete dispatch error", zap.Error(err))
			}
		}
	}()
}
