// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: github.com/micro/quota-srv/proto/quota.proto

/*
Package quota is a generated protocol buffer package.

It is generated from these files:
	github.com/micro/quota-srv/proto/quota.proto

It has these top-level messages:
	AllocateRequest
	AllocateResponse
	Config
	Allocation
	Update
*/
package quota

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
	context "context"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for Quota service

type QuotaService interface {
	Allocate(ctx context.Context, in *AllocateRequest, opts ...client.CallOption) (*AllocateResponse, error)
}

type quotaService struct {
	c           client.Client
	serviceName string
}

func QuotaServiceClient(serviceName string, c client.Client) QuotaService {
	if c == nil {
		c = client.NewClient()
	}
	if len(serviceName) == 0 {
		serviceName = "quota"
	}
	return &quotaService{
		c:           c,
		serviceName: serviceName,
	}
}

func (c *quotaService) Allocate(ctx context.Context, in *AllocateRequest, opts ...client.CallOption) (*AllocateResponse, error) {
	req := c.c.NewRequest(c.serviceName, "Quota.Allocate", in)
	out := new(AllocateResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Quota service

type QuotaHandler interface {
	Allocate(context.Context, *AllocateRequest, *AllocateResponse) error
}

func RegisterQuotaHandler(s server.Server, hdlr QuotaHandler, opts ...server.HandlerOption) {
	s.Handle(s.NewHandler(&Quota{hdlr}, opts...))
}

type Quota struct {
	QuotaHandler
}

func (h *Quota) Allocate(ctx context.Context, in *AllocateRequest, out *AllocateResponse) error {
	return h.QuotaHandler.Allocate(ctx, in, out)
}