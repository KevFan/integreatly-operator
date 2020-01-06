// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"./pkg/apis/integreatly/v1alpha1.BlobStorage":             schema_pkg_apis_integreatly_v1alpha1_BlobStorage(ref),
		"./pkg/apis/integreatly/v1alpha1.BlobStorageSpec":         schema_pkg_apis_integreatly_v1alpha1_BlobStorageSpec(ref),
		"./pkg/apis/integreatly/v1alpha1.BlobStorageStatus":       schema_pkg_apis_integreatly_v1alpha1_BlobStorageStatus(ref),
		"./pkg/apis/integreatly/v1alpha1.Postgres":                schema_pkg_apis_integreatly_v1alpha1_Postgres(ref),
		"./pkg/apis/integreatly/v1alpha1.PostgresSpec":            schema_pkg_apis_integreatly_v1alpha1_PostgresSpec(ref),
		"./pkg/apis/integreatly/v1alpha1.PostgresStatus":          schema_pkg_apis_integreatly_v1alpha1_PostgresStatus(ref),
		"./pkg/apis/integreatly/v1alpha1.Redis":                   schema_pkg_apis_integreatly_v1alpha1_Redis(ref),
		"./pkg/apis/integreatly/v1alpha1.RedisSpec":               schema_pkg_apis_integreatly_v1alpha1_RedisSpec(ref),
		"./pkg/apis/integreatly/v1alpha1.RedisStatus":             schema_pkg_apis_integreatly_v1alpha1_RedisStatus(ref),
		"./pkg/apis/integreatly/v1alpha1.SMTPCredentialSet":       schema_pkg_apis_integreatly_v1alpha1_SMTPCredentialSet(ref),
		"./pkg/apis/integreatly/v1alpha1.SMTPCredentialSetSpec":   schema_pkg_apis_integreatly_v1alpha1_SMTPCredentialSetSpec(ref),
		"./pkg/apis/integreatly/v1alpha1.SMTPCredentialSetStatus": schema_pkg_apis_integreatly_v1alpha1_SMTPCredentialSetStatus(ref),
	}
}

func schema_pkg_apis_integreatly_v1alpha1_BlobStorage(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BlobStorage is the Schema for the blobstorages API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/integreatly/v1alpha1.BlobStorageSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/integreatly/v1alpha1.BlobStorageStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/integreatly/v1alpha1.BlobStorageSpec", "./pkg/apis/integreatly/v1alpha1.BlobStorageStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_BlobStorageSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BlobStorageSpec defines the desired state of BlobStorage",
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"tier": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"skipCreate": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"secretRef": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"),
						},
					},
				},
				Required: []string{"type", "tier", "secretRef"},
			},
		},
		Dependencies: []string{
			"github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_BlobStorageStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BlobStorageStatus defines the observed state of BlobStorage",
				Properties: map[string]spec.Schema{
					"strategy": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"provider": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"secretRef": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"),
						},
					},
					"phase": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"message": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_Postgres(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Postgres is the Schema for the postgres API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/integreatly/v1alpha1.PostgresSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/integreatly/v1alpha1.PostgresStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/integreatly/v1alpha1.PostgresSpec", "./pkg/apis/integreatly/v1alpha1.PostgresStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_PostgresSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "PostgresSpec defines the desired state of Postgres",
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"tier": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"skipCreate": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"secretRef": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"),
						},
					},
				},
				Required: []string{"type", "tier", "secretRef"},
			},
		},
		Dependencies: []string{
			"github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_PostgresStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "PostgresStatus defines the observed state of Postgres",
				Properties: map[string]spec.Schema{
					"strategy": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"provider": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"secretRef": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"),
						},
					},
					"phase": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"message": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_Redis(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Redis is the Schema for the redis API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/integreatly/v1alpha1.RedisSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/integreatly/v1alpha1.RedisStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/integreatly/v1alpha1.RedisSpec", "./pkg/apis/integreatly/v1alpha1.RedisStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_RedisSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "RedisSpec defines the desired state of Redis",
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"tier": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"skipCreate": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"secretRef": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"),
						},
					},
				},
				Required: []string{"type", "tier", "secretRef"},
			},
		},
		Dependencies: []string{
			"github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_RedisStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "RedisStatus defines the observed state of Redis",
				Properties: map[string]spec.Schema{
					"strategy": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"provider": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"secretRef": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"),
						},
					},
					"phase": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"message": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_SMTPCredentialSet(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SMTPCredentials is the Schema for the smtpcredentialset API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/integreatly/v1alpha1.SMTPCredentialSetSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/integreatly/v1alpha1.SMTPCredentialSetStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/integreatly/v1alpha1.SMTPCredentialSetSpec", "./pkg/apis/integreatly/v1alpha1.SMTPCredentialSetStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_SMTPCredentialSetSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SMTPCredentialsSpec defines the desired state of SMTPCredentials",
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"tier": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"skipCreate": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"secretRef": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"),
						},
					},
				},
				Required: []string{"type", "tier", "secretRef"},
			},
		},
		Dependencies: []string{
			"github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"},
	}
}

func schema_pkg_apis_integreatly_v1alpha1_SMTPCredentialSetStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SMTPCredentialsStatus defines the observed state of SMTPCredentials",
				Properties: map[string]spec.Schema{
					"strategy": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"provider": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"secretRef": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"),
						},
					},
					"phase": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"message": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/integr8ly/cloud-resource-operator/pkg/apis/integreatly/v1alpha1/types.SecretRef"},
	}
}
