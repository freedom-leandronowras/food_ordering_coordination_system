package stacks

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func StartEC2InstanceStack(ctx *pulumi.Context) error {
	cfg := config.New(ctx, "")

	instanceID := strings.TrimSpace(cfg.Require("ec2InstanceId"))
	awsRegion := strings.TrimSpace(cfg.Get("awsRegion"))
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	commandScript := fmt.Sprintf(`
set -euo pipefail
INSTANCE_ID=%q
REGION=%q

state="$(AWS_CLI_AUTO_PROMPT=off AWS_PAGER="" aws --no-cli-pager ec2 describe-instances --instance-ids "$INSTANCE_ID" --region "$REGION" --query 'Reservations[0].Instances[0].State.Name' --output text)"
if [[ "$state" != "running" ]]; then
  AWS_CLI_AUTO_PROMPT=off AWS_PAGER="" aws --no-cli-pager ec2 start-instances --instance-ids "$INSTANCE_ID" --region "$REGION" >/dev/null
  AWS_CLI_AUTO_PROMPT=off AWS_PAGER="" aws --no-cli-pager ec2 wait instance-running --instance-ids "$INSTANCE_ID" --region "$REGION"
fi

AWS_CLI_AUTO_PROMPT=off AWS_PAGER="" aws --no-cli-pager ec2 describe-instances --instance-ids "$INSTANCE_ID" --region "$REGION" --query 'Reservations[0].Instances[0].{State:State.Name,PublicIp:PublicIpAddress,PrivateIp:PrivateIpAddress,InstanceId:InstanceId}' --output json
`, instanceID, awsRegion)

	startCmd, err := local.NewCommand(ctx, "start-ec2-instance", &local.CommandArgs{
		Create: pulumi.String(commandScript),
		Update: pulumi.String(commandScript),
	})
	if err != nil {
		return err
	}

	ctx.Export("stackMode", pulumi.String("ec2-start"))
	ctx.Export("instanceId", pulumi.String(instanceID))
	ctx.Export("instanceInfo", startCmd.Stdout)

	return nil
}
