package k8s

import (
	// v1alpha1 "github.com/hicompute/histack/api/v1alpha1"
	"github.com/hicompute/kloudstack/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

var Scheme = runtime.NewScheme()

func init() {
	_ = v1alpha1.AddToScheme(Scheme)
	_ = corev1.AddToScheme(Scheme)
	_ = kubevirtv1.AddToScheme(Scheme)
	// add other APIs here (Core, Apps, CRDs, etc.)
}
