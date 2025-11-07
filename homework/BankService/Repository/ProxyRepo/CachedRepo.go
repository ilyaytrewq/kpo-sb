package proxyrepo

import (
	"context"
	"sync"

	repository "github.com/ilyaytrewq/kpo-sb/homework/BankService/Repository"
	service "github.com/ilyaytrewq/kpo-sb/homework/BankService/Service"
)

type CachedRepo struct {
	db    repository.ICommonRepo
	mu    sync.RWMutex
	cache map[service.ObjectID]service.ICommonObject
}

func NewCachedRepo(ctx context.Context, db repository.ICommonRepo) (*CachedRepo, error) {
	all, err := db.All(ctx)
	if err != nil {
		return nil, err
	}
	c := make(map[service.ObjectID]service.ICommonObject, len(all))
	for _, o := range all {
		c[o.ID()] = o
	}
	return &CachedRepo{db: db, cache: c}, nil
}

func (p *CachedRepo) ByID(ctx context.Context, id service.ObjectID) (service.ICommonObject, error) {
	p.mu.RLock()
	obj, ok := p.cache[id]
	p.mu.RUnlock()
	if ok {
		return obj, nil
	}
	// Optional read-through
	o, err := p.db.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.cache[id] = o
	p.mu.Unlock()
	return o, nil
}

func (p *CachedRepo) All(ctx context.Context) ([]service.ICommonObject, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]service.ICommonObject, 0, len(p.cache))
	for _, o := range p.cache {
		out = append(out, o)
	}
	return out, nil
}

func (p *CachedRepo) Save(ctx context.Context, obj service.ICommonObject) error {
	if err := p.db.Save(ctx, obj); err != nil {
		return err
	}
	p.mu.Lock()
	p.cache[obj.ID()] = obj
	p.mu.Unlock()
	return nil
}

func (p *CachedRepo) Delete(ctx context.Context, id service.ObjectID) error {
	if err := p.db.Delete(ctx, id); err != nil {
		return err
	}
	p.mu.Lock()
	delete(p.cache, id)
	p.mu.Unlock()
	return nil
}
