package stacks

import (
	"fmt"

	"food_ordering_coordination_system/internal/constructs"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func DeployLocalStack(ctx *pulumi.Context) error {
	cfg := config.New(ctx, "")

	namespaceName := cfg.Get("namespace")
	if namespaceName == "" {
		namespaceName = "food-ordering"
	}

	imagePullPolicy := cfg.Get("imagePullPolicy")
	if imagePullPolicy == "" {
		imagePullPolicy = "IfNotPresent"
	}

	apiImage := cfg.Get("apiImage")
	if apiImage == "" {
		apiImage = "food-coordination-api:prod"
	}

	webImage := cfg.Get("webImage")
	if webImage == "" {
		webImage = "food-web-ui:prod"
	}

	vendorImage := cfg.Get("vendorImage")
	if vendorImage == "" {
		vendorImage = "food-vendor-mocks:prod"
	}

	mongoImage := cfg.Get("mongoImage")
	if mongoImage == "" {
		mongoImage = "mongo:7"
	}

	webAPIBaseURL := cfg.Get("webApiBaseUrl")
	if webAPIBaseURL == "" {
		webAPIBaseURL = "http://127.0.0.1:18081"
	}

	jwtSigningKey := cfg.RequireSecret("jwtSigningKey")

	allowedEmailDomains := cfg.Get("allowedEmailDomains")
	authAllowSelfAssignRoles := cfg.GetBool("authAllowSelfAssignRoles")

	namespace, err := corev1.NewNamespace(ctx, namespaceName, &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(namespaceName),
		},
	})
	if err != nil {
		return err
	}

	apiSecretName := "coordination-api-secrets"
	apiSecrets, err := corev1.NewSecret(ctx, apiSecretName, &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: pulumi.String(namespaceName),
			Name:      pulumi.String(apiSecretName),
		},
		Type: pulumi.String("Opaque"),
		StringData: pulumi.StringMap{
			"JWT_SIGNING_KEY": jwtSigningKey,
		},
	}, pulumi.DependsOn([]pulumi.Resource{namespace}))
	if err != nil {
		return err
	}

	mongo, err := constructs.NewMongoDB(ctx, constructs.MongoDBArgs{
		Namespace:       namespaceName,
		Name:            "mongodb",
		Image:           mongoImage,
		Replicas:        1,
		ImagePullPolicy: imagePullPolicy,
		DependsOn:       []pulumi.Resource{namespace},
	})
	if err != nil {
		return err
	}

	pizzaVendor, err := constructs.NewWorkload(ctx, constructs.WorkloadArgs{
		Namespace:       namespaceName,
		Name:            "vendor-pizza",
		Image:           vendorImage,
		ContainerPort:   4010,
		ServicePort:     4010,
		Replicas:        1,
		ImagePullPolicy: imagePullPolicy,
		ServiceType:     "ClusterIP",
		Env: map[string]pulumi.StringInput{
			"VENDOR_API_PORT":    pulumi.String("4010"),
			"VENDOR_API_DB_PATH": pulumi.String("/app/pizza_place_db.json"),
		},
		ReadinessPath: "/menu",
		LivenessPath:  "/menu",
		DependsOn:     []pulumi.Resource{namespace},
	})
	if err != nil {
		return err
	}

	sushiVendor, err := constructs.NewWorkload(ctx, constructs.WorkloadArgs{
		Namespace:       namespaceName,
		Name:            "vendor-sushi",
		Image:           vendorImage,
		ContainerPort:   4010,
		ServicePort:     4010,
		Replicas:        1,
		ImagePullPolicy: imagePullPolicy,
		ServiceType:     "ClusterIP",
		Env: map[string]pulumi.StringInput{
			"VENDOR_API_PORT":    pulumi.String("4010"),
			"VENDOR_API_DB_PATH": pulumi.String("/app/sushi_bar_db.json"),
		},
		ReadinessPath: "/menu",
		LivenessPath:  "/menu",
		DependsOn:     []pulumi.Resource{namespace},
	})
	if err != nil {
		return err
	}

	tacoVendor, err := constructs.NewWorkload(ctx, constructs.WorkloadArgs{
		Namespace:       namespaceName,
		Name:            "vendor-taco",
		Image:           vendorImage,
		ContainerPort:   4010,
		ServicePort:     4010,
		Replicas:        1,
		ImagePullPolicy: imagePullPolicy,
		ServiceType:     "ClusterIP",
		Env: map[string]pulumi.StringInput{
			"VENDOR_API_PORT":    pulumi.String("4010"),
			"VENDOR_API_DB_PATH": pulumi.String("/app/taco_truck_db.json"),
		},
		ReadinessPath: "/menu",
		LivenessPath:  "/menu",
		DependsOn:     []pulumi.Resource{namespace},
	})
	if err != nil {
		return err
	}

	vendorURLs := fmt.Sprintf(
		"a1b2c3d4-1111-4000-a000-000000000001=Bella Napoli=http://vendor-pizza:4010," +
			"a1b2c3d4-2222-4000-a000-000000000002=Sakura Sushi=http://vendor-sushi:4010," +
			"a1b2c3d4-3333-4000-a000-000000000003=El Fuego Taco=http://vendor-taco:4010",
	)

	api, err := constructs.NewWorkload(ctx, constructs.WorkloadArgs{
		Namespace:       namespaceName,
		Name:            "coordination-api",
		Image:           apiImage,
		ContainerPort:   8080,
		ServicePort:     8080,
		Replicas:        1,
		ImagePullPolicy: imagePullPolicy,
		ServiceType:     "ClusterIP",
		Env: map[string]pulumi.StringInput{
			"PORT":                         pulumi.String("8080"),
			"MONGODB_URI":                  pulumi.String("mongodb://mongodb:27017"),
			"MONGODB_DATABASE":             pulumi.String("food_ordering"),
			"VENDOR_URLS":                  pulumi.String(vendorURLs),
			"AUTH_ALLOW_SELF_ASSIGN_ROLES": pulumi.String(fmt.Sprintf("%t", authAllowSelfAssignRoles)),
		},
		SecretEnv: map[string]constructs.SecretEnvVar{
			"JWT_SIGNING_KEY": {
				SecretName: apiSecretName,
				Key:        "JWT_SIGNING_KEY",
			},
		},
		ReadinessPath: "/api/vendors",
		LivenessPath:  "/api/vendors",
		DependsOn: []pulumi.Resource{
			mongo.Deployment,
			pizzaVendor.Deployment,
			sushiVendor.Deployment,
			tacoVendor.Deployment,
			apiSecrets,
		},
	})
	if err != nil {
		return err
	}

	webEnv := map[string]pulumi.StringInput{
		"PORT":                     pulumi.String("3000"),
		"HOST":                     pulumi.String("0.0.0.0"),
		"HOSTNAME":                 pulumi.String("0.0.0.0"),
		"NEXT_PUBLIC_API_BASE_URL": pulumi.String(webAPIBaseURL),
	}
	if allowedEmailDomains != "" {
		webEnv["NEXT_PUBLIC_ALLOWED_EMAIL_DOMAINS"] = pulumi.String(allowedEmailDomains)
	}

	web, err := constructs.NewWorkload(ctx, constructs.WorkloadArgs{
		Namespace:       namespaceName,
		Name:            "web-ui",
		Image:           webImage,
		ContainerPort:   3000,
		ServicePort:     3000,
		Replicas:        1,
		ImagePullPolicy: imagePullPolicy,
		ServiceType:     "ClusterIP",
		Env:             webEnv,
		ReadinessPath:   "/auth",
		LivenessPath:    "/auth",
		DependsOn: []pulumi.Resource{
			api.Deployment,
		},
	})
	if err != nil {
		return err
	}

	ctx.Export("namespace", pulumi.String(namespaceName))
	ctx.Export("apiService", api.Service.Metadata.Name())
	ctx.Export("webService", web.Service.Metadata.Name())
	ctx.Export("vendorServices", pulumi.All(
		pizzaVendor.Service.Metadata.Name(),
		sushiVendor.Service.Metadata.Name(),
		tacoVendor.Service.Metadata.Name(),
	).ApplyT(func(values []interface{}) []string {
		names := make([]string, 0, len(values))
		for _, value := range values {
			if value == nil {
				continue
			}
			name, ok := value.(string)
			if !ok || name == "" {
				continue
			}
			names = append(names, name)
		}
		return names
	}))
	ctx.Export("localAccessInstructions", pulumi.String("Use kubectl port-forward svc/web-ui 18080:3000 and svc/coordination-api 18081:8080"))

	return nil
}
