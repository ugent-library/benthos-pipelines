package main

import (
	"context"

	"github.com/benthosdev/benthos/v4/public/service"

	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/mongodb"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/ugent-library/projects/benthos/processor"
)

func main() {
	service.RunCLI(context.Background())
}
