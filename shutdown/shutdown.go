package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Println("shutting down, press Ctrl+C again to force")
		stop() // unregister: next SIGINT uses the default handler and kills the process
	}()

	var i atomic.Int64
	err := Run(ctx, 5, 5*time.Second, func(ctx context.Context) (Message, error) {
		return Message{int(i.Add(1))}, nil
	}, func(ctx context.Context, m Message) error {
		select {
		case <-time.After(time.Duration(m.ID) * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
		log.Println("job processed:", m.ID)
		return nil
	})
	switch {
	case err == nil:
		log.Println("drained cleanly")
	case errors.Is(err, ErrDrainTimeout):
		log.Println("drain timed out, work was aborted")
		os.Exit(1)
	default:
		log.Println("run:", err)
		os.Exit(1)
	}
}

var ErrDrainTimeout = errors.New("shutdown: drain timeout")

type Message struct{ ID int }

type Source func(ctx context.Context) (Message, error)

func Run(stopCtx context.Context, workers int, drain time.Duration, next Source, handle func(context.Context, Message) error) error {
	workCtx, cancel := context.WithCancel(context.WithoutCancel(stopCtx))
	defer cancel()

	var eg errgroup.Group
	for range workers {
		eg.Go(func() error {
			for {
				if stopCtx.Err() != nil {
					return nil
				}
				m, err := next(stopCtx)
				if err != nil {
					if stopCtx.Err() != nil {
						return nil
					}
					return err
				}
				if err := handle(workCtx, m); err != nil {
					return err
				}
			}
		})
	}

	cErr := make(chan error)

	go func() {
		cErr <- eg.Wait()
	}()

	select {
	case err := <-cErr:
		return err
	case <-stopCtx.Done():
	}

	select {
	case err := <-cErr:
		return err
	case <-time.After(drain):
		cancel()
		<-cErr
		return ErrDrainTimeout
	}
}
