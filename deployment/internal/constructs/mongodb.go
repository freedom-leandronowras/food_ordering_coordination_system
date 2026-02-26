package constructs

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type MongoDBArgs struct {
	Namespace       string
	Name            string
	Image           string
	ImagePullPolicy string
	Replicas        int
	DependsOn       []pulumi.Resource
}

func NewMongoDB(ctx *pulumi.Context, args MongoDBArgs) (*Workload, error) {
	if args.Name == "" {
		args.Name = "mongodb"
	}
	if args.Image == "" {
		args.Image = "mongo:7"
	}

	volumes := corev1.VolumeArray{
		corev1.VolumeArgs{
			Name: pulumi.String("mongo-data"),
			EmptyDir: &corev1.EmptyDirVolumeSourceArgs{
				SizeLimit: pulumi.String("2Gi"),
			},
		},
	}

	mounts := corev1.VolumeMountArray{
		corev1.VolumeMountArgs{
			Name:      pulumi.String("mongo-data"),
			MountPath: pulumi.String("/data/db"),
		},
	}

	return NewWorkload(ctx, WorkloadArgs{
		Namespace:       args.Namespace,
		Name:            args.Name,
		Image:           args.Image,
		ContainerPort:   27017,
		ServicePort:     27017,
		Replicas:        args.Replicas,
		ImagePullPolicy: args.ImagePullPolicy,
		ServiceType:     "ClusterIP",
		Volumes:         volumes,
		VolumeMounts:    mounts,
		DependsOn:       args.DependsOn,
	})
}
