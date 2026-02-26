package main

import (
	"food_ordering_coordination_system/internal/stacks"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		stackMode := strings.TrimSpace(cfg.Get("stackMode"))
		if stackMode == "ec2-start" {
			return stacks.StartEC2InstanceStack(ctx)
		}
		return stacks.DeployLocalStack(ctx)
	})
}
