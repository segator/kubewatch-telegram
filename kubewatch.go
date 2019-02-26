package main

//-----------------------------------------------------------------------------
// Package factored import statement:
//-----------------------------------------------------------------------------

import (

	// Stdlib:
	// "encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	// Telegram
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	// Kubernetes:
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	// Community:
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

//-----------------------------------------------------------------------------
// Command, flags and arguments:
//-----------------------------------------------------------------------------

var (
	// Root level command:
	app = kingpin.New("kubewatch", "Watches Kubernetes resources via its API.")

	// Resources:
	resources = []string{
		"configMaps", "endpoints", "events", "limitranges",
		"persistentvolumeclaims", "persistentvolumes", "pods", "podtemplates",
		"replicationcontrollers", "resourcequotas", "secrets", "serviceaccounts",
		"services", "deployments", "horizontalpodautoscalers", "ingresses", "jobs"}

	// Flags:
	flgKubeconfig = app.Flag("kubeconfig",
		"Absolute path to the kubeconfig file.").
		Default(kubeconfigPath()).ExistingFileOrDir()

	flgNamespace = app.Flag("namespace",
		"Set the namespace to be watched.").
		Default(v1.NamespaceAll).HintAction(listNamespaces).String()

	flgFlatten = app.Flag("flatten",
		"Whether to produce flatten JSON output or not.").Bool()

	argAPI = app.Flag("telegramapi",
		"API Key for Telegram bot").Required().String()

	argGroup = app.Flag("telegramgroup",
		`Group that the bot should post to. 
Note that Telegram groups are negative values, but drop the - here. 
If you wish to message an individual, you will need to add a negative on the command line`).Required().Int64()
	// Arguments:
	argResources = app.Arg("resources",
		"Space delimited list of resources to be watched.").
		Required().HintOptions(resources...).Enums(resources...)
)

//-----------------------------------------------------------------------------
// Types and structs:
//-----------------------------------------------------------------------------

type verObj struct {
	apiVersion    string
	runtimeObject runtime.Object
}

type strIfce map[string]interface{}

//-----------------------------------------------------------------------------
// Map resources to runtime objects:
//-----------------------------------------------------------------------------

var resourceObject = map[string]verObj{

	// v1:
	"configMaps":             {"v1", &v1.ConfigMap{}},
	"endpoints":              {"v1", &v1.Endpoints{}},
	"events":                 {"v1", &v1.Event{}},
	"limitranges":            {"v1", &v1.LimitRange{}},
	"persistentvolumeclaims": {"v1", &v1.PersistentVolumeClaim{}},
	"persistentvolumes":      {"v1", &v1.PersistentVolume{}},
	"pods":                   {"v1", &v1.Pod{}},
	"podtemplates":           {"v1", &v1.PodTemplate{}},
	"replicationcontrollers": {"v1", &v1.ReplicationController{}},
	"resourcequotas":         {"v1", &v1.ResourceQuota{}},
	"secrets":                {"v1", &v1.Secret{}},
	"serviceaccounts":        {"v1", &v1.ServiceAccount{}},
	"services":               {"v1", &v1.Service{}},

	// v1beta1:
	"deployments":              {"v1beta1", &v1beta1.Deployment{}},
	"horizontalpodautoscalers": {"v1beta1", &v1beta1.HorizontalPodAutoscaler{}},
	"ingresses":                {"v1beta1", &v1beta1.Ingress{}},
	"jobs":                     {"v1beta1", &v1beta1.Job{}},
}

//-----------------------------------------------------------------------------
// func init() is called after all the variable declarations in the package
// have evaluated their initializers, and those are evaluated only after all
// the imported packages have been initialized:
//-----------------------------------------------------------------------------

func init() {

	// Customize kingpin:
	app.Version("v0.3.2").Author("Marc Villacorta Morera")
	app.UsageTemplate(usageTemplate)
	app.HelpFlag.Short('h')

	// Customize the default logger:
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
}

//-----------------------------------------------------------------------------
// Entry point:
//-----------------------------------------------------------------------------

func main() {

	// Parse command flags:
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Build the config:
	config, err := buildConfig(*flgKubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create the clientset:
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Watch for the given resource:
	for _, resource := range *argResources {
		watchResource(clientset, resource, *flgNamespace)
	}

	// Block forever:
	select {}
}

//-----------------------------------------------------------------------------
// watchResource:
//-----------------------------------------------------------------------------

func watchResource(clientset *kubernetes.Clientset, resource, namespace string) {

	var client rest.Interface

	// Set the API endpoint:
	switch resourceObject[resource].apiVersion {
	case "v1":
		client = clientset.Core().RESTClient()
	case "v1beta1":
		client = clientset.Extensions().RESTClient()
	}

	// Watch for resource in namespace:
	listWatch := cache.NewListWatchFromClient(
		client, resource, namespace,
		fields.Everything())

	// Ugly hack to suppress sync events:
	listWatch.ListFunc = func(options api.ListOptions) (runtime.Object, error) {
		return client.Get().Namespace("none").Resource(resource).Do().Get()
	}

	// Controller providing event notifications:
	_, controller := cache.NewInformer(
		listWatch, resourceObject[resource].runtimeObject,
		time.Second*0, cache.ResourceEventHandlerFuncs{
			AddFunc:    outputEvent,
			DeleteFunc: outputEvent,
		},
	)

	// Log this watch:
	log.WithField("type", resource).Info("Watching for new resources")

	// Start the controller:
	go controller.Run(wait.NeverStop)
}

//-----------------------------------------------------------------------------
// printEvent:
//-----------------------------------------------------------------------------

func outputEvent(obj interface{}) {
	bot, err := tgbotapi.NewBotAPI(*argAPI)

	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	// log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	switch v := obj.(type) {
	case *v1.Pod:
		// mObj := v.(*v1.Pod)
		printPod(v, bot)
	}

}

func printPod(v *v1.Pod, bot *tgbotapi.BotAPI) {
	group := int64(-1) * *argGroup

	if len(v.OwnerReferences) > 0 {
		msg := tgbotapi.NewMessage(group, fmt.Sprintf("Pod: %s changed state to %s\n", v.OwnerReferences[0].Name, v.Status.Phase))
		bot.Send(msg)

	} else {
		msg := tgbotapi.NewMessage(group, fmt.Sprintf("Pod: %s changed state to %s\n", v.GetName(), v.Status.Phase))
		bot.Send(msg)
	}
}

//-----------------------------------------------------------------------------
// kubeconfigPath:
//-----------------------------------------------------------------------------

func kubeconfigPath() (path string) {

	// Return ~/.kube/config if exists...
	if _, err := os.Stat(os.Getenv("HOME") + "/.kube/config"); err == nil {
		return os.Getenv("HOME") + "/.kube/config"
	}

	// ...otherwise return '.':
	return "."
}

//-----------------------------------------------------------------------------
// buildConfig:
//-----------------------------------------------------------------------------

func buildConfig(kubeconfig string) (*rest.Config, error) {

	// Use kubeconfig if given...
	if kubeconfig != "" && kubeconfig != "." {

		// Log and return:
		log.WithField("file", kubeconfig).Info("Running out-of-cluster using kubeconfig")
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// ...otherwise assume in-cluster:
	log.Info("Running in-cluster using environment variables")
	return rest.InClusterConfig()
}

//-----------------------------------------------------------------------------
// listNamespaces:
//-----------------------------------------------------------------------------

func listNamespaces() (list []string) {

	// Build the config:
	config, err := buildConfig(*flgKubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create the clientset:
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Get the list of namespace objects:
	l, err := clientset.Namespaces().List(v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Extract the name of each namespace:
	for _, v := range l.Items {
		list = append(list, v.Name)
	}

	return
}

//-----------------------------------------------------------------------------
// flatten:
//-----------------------------------------------------------------------------

func flatten(r strIfce, p string, v reflect.Value) {

	// Append '_' to prefix:
	if p != "" {
		p = p + "_"
	}

	// Set the value:
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	// Return if !valid:
	if !v.IsValid() {
		return
	}

	// Set the type:
	t := v.Type()

	// Flatten each type kind:
	switch t.Kind() {
	case reflect.Bool:
		flattenBool(v, p, r)
	case reflect.Float64:
		flattenFloat64(v, p, r)
	case reflect.Map:
		flattenMap(v, p, r)
	case reflect.Slice:
		flattenSlice(v, p, r)
	case reflect.String:
		flattenString(v, p, r)
	default:
		log.Error("Unknown: " + p)
	}
}

//-----------------------------------------------------------------------------
// flattenBool:
//-----------------------------------------------------------------------------

func flattenBool(v reflect.Value, p string, r strIfce) {
	if v.Bool() {
		r[p[:len(p)-1]] = "true"
	} else {
		r[p[:len(p)-1]] = "false"
	}
}

//-----------------------------------------------------------------------------
// flattenFloat64:
//-----------------------------------------------------------------------------

func flattenFloat64(v reflect.Value, p string, r strIfce) {
	r[p[:len(p)-1]] = fmt.Sprintf("%f", v.Float())
}

//-----------------------------------------------------------------------------
// flattenMap:
//-----------------------------------------------------------------------------

func flattenMap(v reflect.Value, p string, r strIfce) {
	for _, k := range v.MapKeys() {
		if k.Kind() == reflect.Interface {
			k = k.Elem()
		}
		if k.Kind() != reflect.String {
			log.Errorf("%s: map key is not string: %s", p, k)
		}
		flatten(r, p+k.String(), v.MapIndex(k))
	}
}

//-----------------------------------------------------------------------------
// flattenSlice:
//-----------------------------------------------------------------------------

func flattenSlice(v reflect.Value, p string, r strIfce) {
	r[p+"#"] = fmt.Sprintf("%d", v.Len())
	for i := 0; i < v.Len(); i++ {
		flatten(r, fmt.Sprintf("%s%d", p, i), v.Index(i))
	}
}

//-----------------------------------------------------------------------------
// flattenString:
//-----------------------------------------------------------------------------

func flattenString(v reflect.Value, p string, r strIfce) {
	r[p[:len(p)-1]] = v.String()
}
