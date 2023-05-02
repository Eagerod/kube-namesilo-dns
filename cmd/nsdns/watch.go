package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/nsdns"
)

func watchCommand() *cobra.Command {
	var ingressClass string
	var domainName string
	var leaseName string

	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "watch ingress objects, and create dns records",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetLevel(log.DebugLevel)

			dm, err := nsdns.NewDnsManager(domainName, ingressClass)
			if err != nil {
				return err
			}

			dm.RefreshesCacheOnUpdate = true
			if err := dm.UpdateCache(); err != nil {
				return err
			}

			go func() {
				for range time.Tick(time.Hour) {
					if err := dm.UpdateCache(); err != nil {
						panic(err)
					}
				}
			}()

			clientset, err := GetKubernetesClientSet()
			if err != nil {
				return err
			}

			var stop chan<- struct{}
			informerFactory := informer(dm, clientset)

			if leaseName != "" {
				// Validate fields present in env
				podNamespace := os.Getenv("POD_NAMESPACE")
				if podNamespace == "" {
					return errors.New("must provide POD_NAMESPACE to use leases")
				}

				podName := os.Getenv("POD_NAME")
				if podName == "" {
					return errors.New("must provide POD_NAME to use leases")
				}

				leConfig := leaderelection.LeaderElectionConfig{
					Lock: &resourcelock.LeaseLock{
						LeaseMeta: metav1.ObjectMeta{
							Name:      leaseName,
							Namespace: podNamespace,
						},
						Client: clientset.CoordinationV1(),
						LockConfig: resourcelock.ResourceLockConfig{
							Identity: podName,
						},
					},
					ReleaseOnCancel: true,
					LeaseDuration:   30 * time.Second,
					RenewDeadline:   15 * time.Second,
					RetryPeriod:     5 * time.Second,
					Callbacks: leaderelection.LeaderCallbacks{
						OnStartedLeading: func(ctx context.Context) {
							stop = runLoop(informerFactory)
						},
						OnStoppedLeading: func() {
							stop <- struct{}{}
						},
					},
				}

				le, err := leaderelection.NewLeaderElector(leConfig)
				if err != nil {
					return err
				}

				le.Run(context.Background())
			} else {
				stop = runLoop(informerFactory)
			}

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

			<-sig
			stop <- struct{}{}

			return nil
		},
	}

	watchCmd.Flags().StringVarP(&ingressClass, "ingress-class", "i", "", "ingress class to use for public DNS records")
	watchCmd.Flags().StringVarP(&domainName, "domain", "d", "", "domain name for API calls")
	watchCmd.Flags().StringVarP(&leaseName, "with-lease", "l", "", "use Kubernetes leases to control leader election")

	return watchCmd
}

func informer(dm *nsdns.DnsManager, clientset *kubernetes.Clientset) informers.SharedInformerFactory {
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)

	informerFactory.Networking().V1().Ingresses().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ingress := obj.(*networkingv1.Ingress)

				if err := dm.HandleIngressExists(ingress); err != nil {
					log.Error(err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				ingress := obj.(*networkingv1.Ingress)

				if err := dm.HandleIngressDeleted(ingress); err != nil {
					log.Error(err)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				ingress := new.(*networkingv1.Ingress)

				if err := dm.HandleIngressExists(ingress); err != nil {
					log.Error(err)
				}
			},
		},
	)

	return informerFactory
}

func runLoop(informerFactory informers.SharedInformerFactory) chan<- struct{} {
	stop := make(chan struct{})
	informerFactory.Start(stop)
	informerFactory.WaitForCacheSync(stop)
	return stop
}
