package gatekeeper

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	k8s "github.com/hicompute/kloudstack/pkg/k8s"
	"github.com/spf13/viper"
	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GateKeepr struct {
	client client.Client
}

func New() *GateKeepr {
	k8sClient, err := k8s.NewClient()
	if err != nil {
		klog.Fatalf("error on creating k8s client: %v", err)
	}
	return &GateKeepr{
		client: k8sClient,
	}
}

func (g *GateKeepr) authorize(w http.ResponseWriter, r *http.Request) {

	var review authv1.SubjectAccessReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, _ := json.MarshalIndent(review, "", "  ")
	klog.Infof("Full SAR request:\n%s", string(body))

	allowed := false
	reason := ""
	nsNamePattern := regexp.MustCompile(`^\d+$`)
	namespace := review.Spec.ResourceAttributes.Namespace
	klog.Infof("Resource Namespace: %s", namespace)
	if !nsNamePattern.MatchString(namespace) {
		klog.Infof("Namespace %s does not flow a number string pattern.", namespace)
		reason = "Namespace " + namespace + " does not flow a number string pattern."
	} else {
		klog.Infof("User: %v", review.Spec.User)
		klog.Infof("Groups: %v", review.Spec.Groups)

		re := regexp.MustCompile(`^[^/]+`)
		for _, group := range review.Spec.Groups {
			match := re.FindString(group)
			klog.Infof("group: %s, match: %s, namespace: %s", group, match, namespace)
			if namespace == match {
				allowed = true
				reason = "User is authenticated."
				break
			}
		}

		ctx := context.Background()
		var ns corev1.Namespace
		err := g.client.Get(ctx, client.ObjectKey{Name: namespace}, &ns)
		if err != nil {
			if errors.IsNotFound(err) {
				ns.Name = namespace
				if err := g.client.Create(ctx, &ns); err != nil {
					klog.Errorf("error on creating namespace %s. %v", namespace, err)
				}
			} else {
				klog.Errorf("error on retrieving namespace %s. %v", namespace, err)
				allowed = false
				reason = "error on retrieving namespace."
			}
		}
	}
	response := authv1.SubjectAccessReview{
		Status: authv1.SubjectAccessReviewStatus{
			Allowed: allowed,
			Reason:  reason,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (g *GateKeepr) Start() error {
	http.HandleFunc("/authorize", g.authorize)
	port := viper.GetString("PORT")
	isTLS := viper.GetBool("TLS")

	if isTLS {
		return http.ListenAndServeTLS(
			":"+port,
			"/tls/tls.crt",
			"/tls/tls.key",
			nil,
		)
	}
	return http.ListenAndServe(":"+port, nil)
}
