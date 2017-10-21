package subscriber

import (
	"github.com/micro/quota-srv/manager"
	proto "github.com/micro/quota-srv/proto"

	"golang.org/x/net/context"
)

func Update(ctx context.Context, u *proto.Update) error {
	manager.Update(u)
	return nil
}
