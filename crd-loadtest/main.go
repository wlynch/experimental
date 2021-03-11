package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/tektoncd/experimental/crd-loadtest/pkg/apis/wlynch.dev/v1alpha1"
	v1alpha1client "github.com/tektoncd/experimental/crd-loadtest/pkg/client/clientset/versioned"
	v1alpha1informer "github.com/tektoncd/experimental/crd-loadtest/pkg/client/informers/externalversions/wlynch.dev/v1alpha1"
	v1alpha1list "github.com/tektoncd/experimental/crd-loadtest/pkg/client/listers/wlynch.dev/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	config.Burst = 10
	config.QPS = 10
	client, err := v1alpha1client.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	informer := v1alpha1informer.NewCRDLoadtestInformer(client, "default", 1*time.Minute, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	stop := make(chan struct{})
	go informer.Run(stop)
	for !informer.HasSynced() {
		fmt.Println("waiting for informer...")
		time.Sleep(1 * time.Second)
	}
	lister := v1alpha1list.NewCRDLoadtestLister(informer.GetIndexer())
	ctx := context.Background()

	/*
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {

			},
		}
	*/

	fmt.Println("N Start Create List Cold Warm")
	for i := 1; i <= 1000000; i++ {
		id := uuid.New().String()
		o := &v1alpha1.CRDLoadtest{
			ObjectMeta: metav1.ObjectMeta{
				Name: id,
				Labels: map[string]string{
					"id": id,
				},
			},
		}

		start := time.Now()

		if _, err := client.WlynchV1alpha1().CRDLoadtests("default").Create(ctx, o, metav1.CreateOptions{}); err != nil {
			fmt.Fprintln(os.Stderr, "Create:", err)
			return
		}
		create := time.Now()

		lr, err := client.WlynchV1alpha1().CRDLoadtests("default").List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("id=%s", id),
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "List:", err)
			return
		}
		list := time.Now()
		_ = lr
		//fmt.Fprintln(os.Stderr, "List:", lr)

		l := labels.SelectorFromSet(o.GetLabels())

		r, err := lister.CRDLoadtests("default").List(l)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cold:", err)
			return
		}
		cold := time.Now()
		_ = r
		//fmt.Fprintln(os.Stderr, "Cold:", r)

		r, err = lister.CRDLoadtests("default").List(l)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warm:", err)
			return
		}
		end := time.Now()
		//fmt.Fprintln(os.Stderr, r)

		fmt.Printf("%d %v %v %v %v %v\n", i, start.Format(time.RFC3339Nano), create.Format(time.RFC3339Nano), list.Format(time.RFC3339Nano), cold.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))
	}
}
