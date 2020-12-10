package sbi

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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

//md5hash gives first 5 bytes of md5 hash of a string. Returns normal ascii string
func md5hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil)[:5])
}

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

	//simulate orchestration delay i.e model selection + fetch  etc.
	time.Sleep(3 * time.Second)
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
	underlayaddr := "edgeunderlay-" + strings.Split(myhostname, "-")[1]
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
	mincli := fmt.Sprint(minclients)
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
	name := "fedserv-" + md5hash(model+source)

	// create Deployment
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"modelkind":  model,
				"sourcekind": source,
			},
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
							Image:   "abh15/flwr:latest",
							Command: []string{"python", "-m", "flwr_example.factory.server", "--mincli", mincli},
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
	log.Println("Creating deployment...")
	depresult, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("Created deployment %q.\n", depresult.GetObjectMeta().GetName())

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
			Selector: map[string]string{
				"modelkind":  model,
				"sourcekind": source,
			},
			Ports: []apiv1.ServicePort{
				{
					Name: "rest",
					Port: 5000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(5000),
					},
				},
				{
					Name: "flower",
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
	log.Println("Creating service...")
	svcresult, err := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("Created service %q.\n", svcresult.GetObjectMeta().GetName())

	// Check if pod is up before initilising
	reqlabel := "modelkind=" + model + ",sourcekind=" + source
	watch, err := clientset.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{LabelSelector: reqlabel})
	if err != nil {
		log.Fatal(err.Error())
	}
	for event := range watch.ResultChan() {
		//fmt.Printf("Type: %v\n", event.Type)
		p, ok := event.Object.(*apiv1.Pod)
		if !ok {
			log.Fatal("unexpected type")
		}
		if p.Status.Phase == "Running" {
			log.Printf("Fed server %v is up", name)
			break
		}
	}
	log.Printf("Fed server requires minimum %v clients", minclients)

	/* 	//Wait for flask to come up
	   	time.Sleep(15 * time.Second)

	   	//Initialse server with min number of clients required to start sampling
	   	data := url.Values{"maxcli": {fmt.Sprint(minclients)}}
	   	resp, err := http.PostForm("http://"+name+":5000/launchserv", data)
	   	if err != nil {
	   		log.Println(err.Error())
	   	}
	   	resp.Body.Close()
	   	log.Println("Server initialised") */

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
	// get svc with required labels and return svcname
	reqlabel := "modelkind=" + model + ",sourcekind=" + source

	svcs, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{LabelSelector: reqlabel})
	if err != nil {
		fmt.Println(err.Error())
	}
	if len(svcs.Items) != 0 {
		var svcname string
		for _, v := range svcs.Items {
			svcname = v.GetName()
		}
		return true, svcname
	}
	return false, ""
}
