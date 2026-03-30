package cni

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"strings"

	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/dynamic"

	"github.com/aws/eks-hybrid/test/e2e/kubernetes"
)

// For kubernetes versions less than 1.30, the cilium template uses
// annonations to add AppArmor configuration
//
//go:embed testdata/cilium/cilium-template-129.yaml
var ciliumTemplate129 []byte

// For kubernetes versions 1.30 and above, the AppArmor configuration
// is in spec.securityContext which is only available in 1.30+
//
//go:embed testdata/cilium/cilium-template-130.yaml
var ciliumTemplate130 []byte

type Cilium struct {
	k8s               dynamic.Interface
	kubernetesVersion string
	// podCIDR is the cluster level CIDR to be use for Pods. It needs to be big enough for
	// Hybrid Nodes.
	//
	// Check the cilium-template file for the node pod cidr mask. The default is 24.
	podCIDR string
	region  string
}

func NewCilium(k8s dynamic.Interface, podCIDR, region, kubernetesVersion string) Cilium {
	return Cilium{
		k8s:               k8s,
		kubernetesVersion: kubernetesVersion,
		podCIDR:           podCIDR,
		region:            region,
	}
}

// isChinaRegion returns true if the region is a China region
func isChinaRegion(region string) bool {
	return strings.HasPrefix(region, "cn-")
}

// getCiliumImageConfig returns the appropriate image repository and tag based on region
func getCiliumImageConfig(region string) (ciliumRepo, operatorRepo, tag string) {
	if isChinaRegion(region) {
		// Use China ECR registry with version 1.19.1-0
		baseRepo := "907723705730.dkr.ecr." + region + ".amazonaws.com.cn"
		return baseRepo + "/cilium/cilium",
			baseRepo + "/cilium/operator-generic",
			"v1.19.1-0"
	}
	// Use public ECR for all other regions with version 1.18.5-0
	return "public.ecr.aws/eks/cilium/cilium",
		"public.ecr.aws/eks/cilium/operator-generic",
		"v1.18.5-0"
}

// Deploy creates or updates the Cilium reosurces.
func (c Cilium) Deploy(ctx context.Context) error {
	ciliumTemplate, err := ciliumTemplate(c.kubernetesVersion)
	if err != nil {
		return err
	}
	tmpl, err := template.New("cilium").Parse(string(ciliumTemplate))
	if err != nil {
		return err
	}

	ciliumRepo, operatorRepo, tag := getCiliumImageConfig(c.region)

	values := map[string]string{
		"PodCIDR":       c.podCIDR,
		"CiliumImage":   ciliumRepo + ":" + tag,
		"OperatorImage": operatorRepo + ":" + tag,
	}
	installation := &bytes.Buffer{}
	err = tmpl.Execute(installation, values)
	if err != nil {
		return err
	}

	objs, err := kubernetes.YamlToUnstructured(installation.Bytes())
	if err != nil {
		return err
	}

	fmt.Printf("Applying cilium installation (region: %s, cilium image: %s, operator image: %s)\n",
		c.region, values["CiliumImage"], values["OperatorImage"])

	return kubernetes.UpsertManifestsWithRetries(ctx, c.k8s, objs)
}

func ciliumTemplate(kubernetesVersion string) ([]byte, error) {
	kubeVersion, err := version.ParseSemantic(kubernetesVersion + ".0")
	if err != nil {
		return nil, fmt.Errorf("parsing version: %v", err)
	}
	if kubeVersion.LessThan(version.MustParseSemantic("1.30.0")) {
		return ciliumTemplate129, nil
	}
	return ciliumTemplate130, nil
}
