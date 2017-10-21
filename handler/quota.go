package handler

import (
	"github.com/micro/go-micro/errors"
	quota "github.com/micro/quota-srv/manager"
	proto "github.com/micro/quota-srv/proto"

	"golang.org/x/net/context"
)

type Quota struct {
	Name string
}

func (q *Quota) Allocate(ctx context.Context, req *proto.AllocateRequest, rsp *proto.AllocateResponse) error {
	a, err := quota.Allocate(req.Resource, req.Bucket, req.Allocation)
	if err == nil {
		rsp.Status = proto.AllocateResponse_OK
		rsp.Allocation = a
		return nil
	}

	switch err {
	case quota.ErrTooManyRequests:
		rsp.Status = proto.AllocateResponse_REJECT_TOO_MANY_REQUESTS
	case quota.ErrServerError:
		rsp.Status = proto.AllocateResponse_REJECT_SERVER_ERROR
	default:
		return errors.InternalServerError(q.Name+".allocate", err.Error())
	}

	return nil
}
