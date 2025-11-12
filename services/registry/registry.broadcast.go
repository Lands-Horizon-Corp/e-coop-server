package registry

import (
	"context"

	"go.uber.org/zap"
)

func (r *Registry[TData, TResponse, TRequest]) OnCreate(context context.Context, entity *TData) {
	go func() {
		<-context.Done()

		if r.service.Logger != nil {
			r.service.Logger.Info("[BROADCAST] Starting OnCreate dispatch",
				zap.String("operation", "CREATE"))
		}

		topics := r.created(entity)
		payload := r.ToModel(entity)

		if r.service.Logger != nil {
			r.service.Logger.Info("[BROADCAST] Dispatching CREATE event",
				zap.Strings("topics", topics))
		}

		if err := r.service.Broker.Dispatch(context, topics, payload); err != nil {
			if r.service.Logger != nil {
				r.service.Logger.Error("[BROADCAST] OnCreate dispatch failed",
					zap.Error(err),
					zap.Strings("topics", topics),
					zap.String("operation", "CREATE"))
			}
		} else {
			if r.service.Logger != nil {
				r.service.Logger.Info("[BROADCAST] OnCreate dispatch successful",
					zap.Strings("topics", topics),
					zap.String("operation", "CREATE"))
			}
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnUpdate(context context.Context, entity *TData) {
	go func() {
		<-context.Done()

		if r.service.Logger != nil {
			r.service.Logger.Info("[BROADCAST] Starting OnUpdate dispatch",
				zap.String("operation", "UPDATE"))
		}

		topics := r.updated(entity)
		payload := r.ToModel(entity)

		if r.service.Logger != nil {
			r.service.Logger.Info("[BROADCAST] Dispatching UPDATE event",
				zap.Strings("topics", topics))
		}

		if err := r.service.Broker.Dispatch(context, topics, payload); err != nil {
			if r.service.Logger != nil {
				r.service.Logger.Error("[BROADCAST] OnUpdate dispatch failed",
					zap.Error(err),
					zap.Strings("topics", topics),
					zap.String("operation", "UPDATE"))
			}
		} else {
			if r.service.Logger != nil {
				r.service.Logger.Info("[BROADCAST] OnUpdate dispatch successful",
					zap.Strings("topics", topics),
					zap.String("operation", "UPDATE"))
			}
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnDelete(context context.Context, entity *TData) {
	go func() {
		<-context.Done()

		if r.service.Logger != nil {
			r.service.Logger.Info("[BROADCAST] Starting OnDelete dispatch",
				zap.String("operation", "DELETE"))
		}

		topics := r.deleted(entity)
		payload := r.ToModel(entity)

		if r.service.Logger != nil {
			r.service.Logger.Info("[BROADCAST] Dispatching DELETE event",
				zap.Strings("topics", topics))
		}

		if err := r.service.Broker.Dispatch(context, topics, payload); err != nil {
			if r.service.Logger != nil {
				r.service.Logger.Error("[BROADCAST] OnDelete dispatch failed",
					zap.Error(err),
					zap.Strings("topics", topics),
					zap.String("operation", "DELETE"))
			}
		} else {
			if r.service.Logger != nil {
				r.service.Logger.Info("[BROADCAST] OnDelete dispatch successful",
					zap.Strings("topics", topics),
					zap.String("operation", "DELETE"))
			}
		}
	}()
}
