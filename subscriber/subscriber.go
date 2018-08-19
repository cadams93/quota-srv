package subscriber

import (
	"github.com/microhq/quota-srv/manager"
	proto "github.com/microhq/quota-srv/proto"

	"golang.org/x/net/context"
)

func Update(ctx context.Context, u *proto.Update) error {
	manager.Update(u)
	return nil
}
