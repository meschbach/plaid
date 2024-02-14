package project

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/resources"
)

type SubresourceNextStep uint8

const (
	SubresourceExists SubresourceNextStep = iota
	SubresourceCreated
)

func (c SubresourceNextStep) String() string {
	switch c {
	case SubresourceCreated:
		return "claimed-created"
	case SubresourceExists:
		return "claimed-exists"
	default:
		return fmt.Sprintf("claimed-step(%d)", c)
	}
}

type Subresource[Status any] struct {
	// Created indicates the claimed resource has been created and Ref is valid.  Clients should not modify this field directly
	Created bool
	// Ref is the resource name which is being watched.  Clients should not modify this field directly
	Ref        resources.Meta
	isWatching bool
	token      resources.WatchToken
}

func (c *Subresource[Status]) Decide(ctx context.Context, env *resourceEnv, status *Status) (SubresourceNextStep, error) {
	if !c.Created {
		return SubresourceCreated, nil
	}

	exists, err := env.rpc.GetStatus(ctx, c.Ref, status)
	if err != nil {
		return SubresourceExists, err
	}
	if !exists {
		c.Created = false
		if err := c.cleanupWatcher(ctx, env); err != nil {
			return SubresourceExists, err
		}
		return SubresourceCreated, nil
	}
	//todo: check status to ensure it is in the correct state
	return SubresourceExists, nil
}

func (c *Subresource[Status]) cleanupWatcher(ctx context.Context, env *resourceEnv) error {
	if !c.isWatching {
		return nil
	}

	if err := env.watcher.Off(ctx, c.token); err != nil {
		return err
	}
	c.isWatching = false
	return nil
}

func (c *Subresource[Status]) Create(ctx context.Context, env *resourceEnv, ref resources.Meta, spec any, opts ...resources.CreateOpt) error {
	token, err := env.watcher.OnResourceStatusChanged(ctx, ref, func(ctx context.Context, changed resources.ResourceChanged) error {
		return env.reconcile(ctx)
	})
	if err != nil {
		return err
	}
	c.isWatching = true
	c.token = token

	if err := env.rpc.Create(ctx, ref, spec, append(opts, resources.ClaimedBy(env.which))...); err != nil {
		if watcherErr := c.cleanupWatcher(ctx, env); watcherErr != nil {
			return errors.Join(watcherErr, err)
		}
		return err
	}
	c.Created = true
	c.Ref = ref
	return nil
}

func (c *Subresource[Status]) Delete(ctx context.Context, env *resourceEnv) error {
	if !c.Created {
		return c.cleanupWatcher(ctx, env)
	}

	cleanUpErr := c.cleanupWatcher(ctx, env)
	_, deleteErr := env.rpc.Delete(ctx, c.Ref)
	c.Created = false
	return errors.Join(cleanUpErr, deleteErr)
}
