package constructs

import (
	"sort"

	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SecretEnvVar struct {
	SecretName string
	Key        string
	Optional   bool
}

type WorkloadArgs struct {
	Namespace       string
	Name            string
	Image           string
	ContainerPort   int
	ServicePort     int
	Replicas        int
	ImagePullPolicy string
	ServiceType     string
	NodePort        int
	Env             map[string]pulumi.StringInput
	SecretEnv       map[string]SecretEnvVar
	ReadinessPath   string
	LivenessPath    string
	Volumes         corev1.VolumeArray
	VolumeMounts    corev1.VolumeMountArray
	DependsOn       []pulumi.Resource
}

type Workload struct {
	Deployment *appsv1.Deployment
	Service    *corev1.Service
}

func NewWorkload(ctx *pulumi.Context, args WorkloadArgs) (*Workload, error) {
	if args.Replicas <= 0 {
		args.Replicas = 1
	}
	if args.ServiceType == "" {
		args.ServiceType = "ClusterIP"
	}
	if args.ImagePullPolicy == "" {
		args.ImagePullPolicy = "IfNotPresent"
	}

	labels := pulumi.StringMap{
		"app.kubernetes.io/name":       pulumi.String(args.Name),
		"app.kubernetes.io/managed-by": pulumi.String("pulumi"),
	}

	container := corev1.ContainerArgs{
		Name:            pulumi.String(args.Name),
		Image:           pulumi.String(args.Image),
		ImagePullPolicy: pulumi.String(args.ImagePullPolicy),
		Ports: corev1.ContainerPortArray{
			corev1.ContainerPortArgs{
				ContainerPort: pulumi.Int(args.ContainerPort),
				Name:          pulumi.String("http"),
			},
		},
		Env:          buildEnvArray(args.Env, args.SecretEnv),
		VolumeMounts: args.VolumeMounts,
	}

	if args.ReadinessPath != "" {
		container.ReadinessProbe = &corev1.ProbeArgs{
			HttpGet: &corev1.HTTPGetActionArgs{
				Path: pulumi.String(args.ReadinessPath),
				Port: pulumi.Int(args.ContainerPort),
			},
			InitialDelaySeconds: pulumi.Int(5),
			PeriodSeconds:       pulumi.Int(10),
			TimeoutSeconds:      pulumi.Int(3),
		}
	}

	if args.LivenessPath != "" {
		container.LivenessProbe = &corev1.ProbeArgs{
			HttpGet: &corev1.HTTPGetActionArgs{
				Path: pulumi.String(args.LivenessPath),
				Port: pulumi.Int(args.ContainerPort),
			},
			InitialDelaySeconds: pulumi.Int(10),
			PeriodSeconds:       pulumi.Int(15),
			TimeoutSeconds:      pulumi.Int(3),
		}
	}

	resourceOpts := []pulumi.ResourceOption{}
	if len(args.DependsOn) > 0 {
		resourceOpts = append(resourceOpts, pulumi.DependsOn(args.DependsOn))
	}

	deployment, err := appsv1.NewDeployment(ctx, args.Name, &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: pulumi.String(args.Namespace),
			Name:      pulumi.String(args.Name),
			Labels:    labels,
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(args.Replicas),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: labels,
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: labels,
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{container},
					Volumes:    args.Volumes,
				},
			},
		},
	}, resourceOpts...)
	if err != nil {
		return nil, err
	}

	servicePort := corev1.ServicePortArgs{
		Name:       pulumi.String("http"),
		Port:       pulumi.Int(args.ServicePort),
		TargetPort: pulumi.Int(args.ContainerPort),
		Protocol:   pulumi.String("TCP"),
	}
	if args.ServiceType == "NodePort" && args.NodePort > 0 {
		servicePort.NodePort = pulumi.Int(args.NodePort)
	}

	service, err := corev1.NewService(ctx, args.Name, &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: pulumi.String(args.Namespace),
			Name:      pulumi.String(args.Name),
			Labels:    labels,
		},
		Spec: &corev1.ServiceSpecArgs{
			Selector: labels,
			Type:     pulumi.String(args.ServiceType),
			Ports: corev1.ServicePortArray{
				servicePort,
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{deployment}))
	if err != nil {
		return nil, err
	}

	return &Workload{Deployment: deployment, Service: service}, nil
}

func buildEnvArray(env map[string]pulumi.StringInput, secretEnv map[string]SecretEnvVar) corev1.EnvVarArray {
	array := corev1.EnvVarArray{}

	literalKeys := make([]string, 0, len(env))
	for key := range env {
		literalKeys = append(literalKeys, key)
	}
	sort.Strings(literalKeys)

	for _, key := range literalKeys {
		array = append(array, corev1.EnvVarArgs{
			Name:  pulumi.String(key),
			Value: env[key],
		})
	}

	secretKeys := make([]string, 0, len(secretEnv))
	for key := range secretEnv {
		secretKeys = append(secretKeys, key)
	}
	sort.Strings(secretKeys)

	for _, key := range secretKeys {
		secret := secretEnv[key]
		array = append(array, corev1.EnvVarArgs{
			Name: pulumi.String(key),
			ValueFrom: &corev1.EnvVarSourceArgs{
				SecretKeyRef: &corev1.SecretKeySelectorArgs{
					Name:     pulumi.String(secret.SecretName),
					Key:      pulumi.String(secret.Key),
					Optional: pulumi.Bool(secret.Optional),
				},
			},
		})
	}

	return array
}
