package main

import (
	"flag"
	"os"
	"os/signal"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog"
)

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {
	klog.InitFlags(nil)

	var kubeconfig string
	var gvrstr string
	// var leaseLockName string
	// var leaseLockNamespace string
	// var id string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&gvrstr, "gvr", "", "the group version resource to be watched")
	// flag.StringVar(&id, "id", "", "the holder identity name")
	// flag.StringVar(&leaseLockName, "lease-lock-name", "example", "the lease lock resource name")
	// flag.StringVar(&leaseLockNamespace, "lease-lock-namespace", "default", "the lease lock resource namespace")
	flag.Parse()

	cfg, err := buildConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("could not get config, err %v", err)
	}

	// Grab a dynamic interface that we can create informers from
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("could not generate dynamic client for config, err %v", err)
	}

	// Create a factory object that we can say "hey, I need to watch this resource"
	// and it will give us back an informer for it
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dc, 0, v1.NamespaceAll, nil)

	// Retrieve a "GroupVersionResource" type that we need when generating our informer from our dynamic factory
	gvr, _ := schema.ParseResourceArg(gvrstr)

	// Finally, create our informer for deployments!
	i := f.ForResource(*gvr)

	stopCh := make(chan struct{})
	go startWatching(stopCh, i.Informer())

	sigCh := make(chan os.Signal, 0)
	signal.Notify(sigCh, os.Kill, os.Interrupt)

	<-sigCh
	close(stopCh)
}

func restConfig() (*rest.Config, error) {
	kubeCfg, err := rest.InClusterConfig()
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		kubeCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		return nil, err
	}

	return kubeCfg, nil
}

func startWatching(stopCh <-chan struct{}, s cache.SharedIndexInformer) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.Info("received add event!")
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			klog.Info("received update event!")
		},
		DeleteFunc: func(obj interface{}) {
			klog.Info("received update event!")
		},
	}

	s.AddEventHandler(handlers)
	s.Run(stopCh)
}
