package app

import (
	"context"

	"github.com/asim/nitro/app/client"
	rpcClient "github.com/asim/nitro/app/client/rpc"
	mevent "github.com/asim/nitro/app/event/memory"
	tmem "github.com/asim/nitro/app/network/memory"
	"github.com/asim/nitro/app/registry/memory"
	"github.com/asim/nitro/app/server"
	rpcServer "github.com/asim/nitro/app/server/rpc"
)

type rpcApp struct {
	opts Options
}

func (s *rpcApp) Name(name string) {
	s.opts.Server.Init(
		server.Name(name),
	)
}

// Init initialises options. Additionally it calls cmd.Init
// which parses command line flags. cmd.Init is only called
// on first Init.
func (s *rpcApp) Init(opts ...Option) {
	// process options
	for _, o := range opts {
		o(&s.opts)
	}
}

func (s *rpcApp) Options() Options {
	return s.opts
}

func (s *rpcApp) Execute(name, ep string, req, rsp interface{}) error {
	r := s.Client().NewRequest(name, ep, req)
	return s.Client().Call(context.Background(), r, rsp)
}

func (s *rpcApp) Broadcast(event string, msg interface{}) error {
	m := s.Client().NewMessage(event, msg)
	return s.Client().Publish(context.Background(), m)
}

func (s *rpcApp) Register(v interface{}) error {
	h := s.Server().NewHandler(v)
	return s.Server().Handle(h)
}

func (s *rpcApp) Subscribe(event string, v interface{}) error {
	sub := s.Server().NewSubscriber(event, v)
	return s.Server().Subscribe(sub)
}

func (s *rpcApp) Client() client.Client {
	return s.opts.Client
}

func (s *rpcApp) Server() server.Server {
	return s.opts.Server
}

func (s *rpcApp) String() string {
	return "rpc"
}

func (s *rpcApp) Start() error {
	for _, fn := range s.opts.BeforeStart {
		if err := fn(); err != nil {
			return err
		}
	}

	if err := s.opts.Server.Start(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStart {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *rpcApp) Stop() error {
	var gerr error

	for _, fn := range s.opts.BeforeStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	if err := s.opts.Server.Stop(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	return gerr
}

func (s *rpcApp) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	// wait on context cancel
	<-s.opts.Context.Done()

	return s.Stop()
}

// New returns a new Nitro app
func New(opts ...Option) *rpcApp {
	b := mevent.NewBroker()
	c := rpcClient.NewClient()
	s := rpcServer.NewServer()
	r := memory.NewRegistry()
	t := tmem.NewTransport()

	// set client options
	c.Init(
		client.Broker(b),
		client.Registry(r),
		client.Transport(t),
	)

	// set server options
	s.Init(
		server.Broker(b),
		server.Registry(r),
		server.Transport(t),
	)

	// define local opts
	options := Options{
		Broker:   b,
		Client:   c,
		Server:   s,
		Registry: r,
		Context:  context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &rpcApp{
		opts: options,
	}
}
