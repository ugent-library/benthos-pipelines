package processor

import (
	"context"
	"encoding/json"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/ugent-library/projects/benthos/cerif"
)

func init() {
	configSpec := service.NewConfigSpec()

	err := service.RegisterProcessor("cerif", configSpec,
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			proc := &cerifProcessor{
				log: mgr.Logger(),
			}

			return proc, nil
		})
	if err != nil {
		panic(err)
	}
}

type cerifProcessor struct {
	log *service.Logger
}

func (c *cerifProcessor) Process(ctx context.Context, m *service.Message) (service.MessageBatch, error) {
	msg, _ := m.AsBytes()

	p, err := cerif.ParseProject(msg)
	if err != nil {
		return nil, err
	}

	k, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	m.SetBytes(k)

	return []*service.Message{m}, nil
}

func (c *cerifProcessor) Close(ctx context.Context) error {
	return nil
}
