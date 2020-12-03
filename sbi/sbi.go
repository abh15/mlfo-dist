package sbi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/abh15/mlfo-dist/parser"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//ResolveRequirements talks with the underlay to match requirements with available resource
func ResolveRequirements(s string, r parser.Requirements) string {
	var resourceID string
	_ = r
	switch s {
	case "source":
		resourceID = "robot.imagedata"
	case "model":
		resourceID = "keras.imagerec"
	case "sink":
		resourceID = "robot.armOptimiser"
	}
	return resourceID
}

//StartFedClients starts fed clients using flwr
func StartFedClients(localsrc string, localmodel string, localsink string, fedIP string, numClients int32) {

	_ = numClients
	data := url.Values{
		"server": {fedIP},
		"source": {localsrc},
		"model":  {localmodel},
		"sink":   {localsink},
	}
	fmt.Printf("\n%+v\n", data)

	//============dynamic service names here
	resp, err := http.PostForm("http://localhost:5000/start", data)

	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	/* var i int32
	for i = 0; i < numClients; i++ {

	} */
}

//StartFedServer starts flwr server
func StartFedServer() string {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	/*
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items)) */
	//=============================================================================
	pods, _ := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{LabelSelector: "app=fedserv"})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fedserv",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "fedserv",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fedserv",
					Labels: map[string]string{
						"app": "fedserv",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    "fedserv",
							Image:   "abh15/flwr:0.91clear",
							Command: []string{"python", "-m", "flwr_example.factory.server"},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 6000,
								},
							},
						},
					}, //
				},
			},
		},
	}
	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	// create service
	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fedserv",
			Labels: map[string]string{
				"app": "fedserv",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []apiv1.ServicePort{
				{
					Protocol: "TCP",
					Port:     6000,
				},
			},
		},
	}

	result2, err2 := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err2 != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result2.GetObjectMeta().GetName())

	return "localhost:8080"
}

func int32Ptr(i int32) *int32 { return &i }
