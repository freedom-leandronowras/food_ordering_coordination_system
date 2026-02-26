package main

import (
	"food_ordering_coordination_system/internal/stacks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		return stacks.DeployLocalStack(ctx)
	})
}
