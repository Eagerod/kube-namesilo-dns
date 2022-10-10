package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func watchCommand() *cobra.Command {
	var ingressClass string
	var domainName string

	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "watch ingress objects, and create dns records",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetLevel(log.DebugLevel)

			if domainName == "" {
				return fmt.Errorf("Must provide a domain name to update DNS records.")
			}

			if ingressClass == "" {
				return fmt.Errorf("Must provide an ingress class to select DNS records.")
			}

			clientset, err := GetKubernetesClientSet()
			if err != nil {
				return err
			}

			informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)

			informerFactory.Networking().V1().Ingresses().Informer().AddEventHandler(
				cache.ResourceEventHandlerFuncs{
					AddFunc: func(obj interface{}) {
						ingress := obj.(*networkingv1.Ingress)
						if !ShouldProcessIngress(ingressClass, ingress) {
							return
						}

						log.Info("Add")
					},
					DeleteFunc: func(obj interface{}) {
						ingress := obj.(*networkingv1.Ingress)
						if !ShouldProcessIngress(ingressClass, ingress) {
							return
						}

						log.Info("Delete")
					},
					UpdateFunc: func(old, new interface{}) {
						ingress := new.(*networkingv1.Ingress)
						if !ShouldProcessIngress(ingressClass, ingress) {
							return
						}

						log.Info("Update")
					},
				},
			)

			stop := make(chan struct{})
			informerFactory.Start(stop)
			informerFactory.WaitForCacheSync(stop)

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

			<-sig

			return nil
		},
	}

	watchCmd.Flags().StringVarP(&ingressClass, "ingress-class", "i", "", "ingress class to use for public DNS records")
	watchCmd.Flags().StringVarP(&domainName, "domain", "d", "", "domain name for API calls")

	return watchCmd
}
