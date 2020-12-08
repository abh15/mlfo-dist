package sbi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/abh15/mlfo-dist/parser"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//int32Ptr is helper function from kubernetes client-go library
func int32Ptr(i int32) *int32 { return &i }

//ResolveRequirements talks with the underlay to match requirements with available resource
func ResolveRequirements(s string, r parser.Requirements) string {
	var resourceID string
	_ = r //should use this intelligently
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

	log.Println("Starting federated clients")

	data := url.Values{
		"server": {fedIP + ":6000"},
		"source": {localsrc},
		"model":  {localmodel},
		"sink":   {localsink},
		"num":    {fmt.Sprint(numClients)},
	}
	myhostname, err := os.Hostname() //Get hostname of this node
	if err != nil {
		log.Println(err.Error())
	}
	underlayaddr := "edgeunderlay" + strings.Split(myhostname, "-")[1]
	resp, err := http.PostForm("http://"+underlayaddr+":5000/startcli", data)
	if err != nil {
		log.Println(err.Error())
	}
	resp.Body.Close()
	log.Println("Started federated clients")

}

//StartFedServer starts a new flwr server and returns the server IP
func StartFedServer(model string, source string, minclients int32) string {
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

	// create new string to use as pod name
	name := "fedserv" + model + source

	// create Deployment
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"modelkind":  model,
					"sourcekind": source,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"modelkind":  model,
						"sourcekind": source,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    name,
							Image:   "abh15/flwr:0.93",
							Command: []string{"python", "-m", "flwr_example.factory.flask-REST"},
							Ports: []apiv1.ContainerPort{
								{ContainerPort: 5000},
								{ContainerPort: 6000},
							},
						},
					},
				},
			},
		},
	}
	// Start Deployment
	fmt.Println("Creating deployment...")
	result, _ := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	log.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	// Start service
	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"modelkind":  model,
				"sourcekind": source,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name: "REST",
					Port: 5000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(5000),
					},
				},
				{
					Name: "Flower",
					Port: 6000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(6000),
					},
				},
			},
		},
	}

	// Deploy service
	result2, _ := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	log.Printf("Created deployment %q.\n", result2.GetObjectMeta().GetName())

	//Initialse server with min number of clients required to start sampling
	data := url.Values{"maxcli": {fmt.Sprint(minclients)}}
	resp, err := http.PostForm("http://"+name+":5000/launchserv", data)
	if err != nil {
		log.Println(err.Error())
	}
	resp.Body.Close()
	log.Println("Server initialised")

	return name
}

//MatchServer checks if fed servers for given model and source already exist
func MatchServer(model string, source string) (bool, string) {
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
	// get pods with required labels
	reqlabel := "modelkind=" + model + ",sourcekind=" + source
	pods, _ := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{LabelSelector: reqlabel})
	if err != nil {
		fmt.Println(err.Error())
	}
	log.Printf("There are %d pods in the cluster\n", len(pods.Items))

	if len(pods.Items) != 0 {
		var svcname string
		for _, v := range pods.Items {
			svcname = strings.Split(v.GetName(), "-")[0]
		}
		return true, svcname
	}
	return false, ""
}
