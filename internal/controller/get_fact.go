package controller

import (
	"context"
	storypb "github.com/NordCoder/Story/generated/api/story"
	"github.com/golang/protobuf/ptypes/empty"
)

func (*implementation) GetFact(ctx context.Context, empty *empty.Empty) (*storypb.GetFactResponse, error) {

}
