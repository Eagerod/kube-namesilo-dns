package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apinetworkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

import (
	"github.com/Eagerod/kube-namesilo-dns/pkg/nsdns"
)

func watchCommand() *cobra.Command {
	var ingressClass string
	var domainName string

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

			clientset, err := GetKubernetesClientSet()
			if err != nil {
				return err
			}

			for err := dm.UpdateCache(); err != nil; {
				log.Error("Initial cache update failed. Retrying in 5 minutes...")
				time.Sleep(5 * time.Minute)
			}

			go func() {
				log.Info("Initial cache update complete. Moving to hourly updates...")
				for {
					time.Sleep(1 * time.Hour)
					for err := dm.UpdateCache(); err != nil; {
						log.Error("Hourly cache update failed. Retrying in 5 minutes...")
						time.Sleep(5 * time.Minute)
					}
				}
			}()

			informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)

			informerFactory.Networking().V1().Ingresses().Informer().AddEventHandler(
				cache.ResourceEventHandlerFuncs{
					AddFunc: func(obj interface{}) {
						ingress := obj.(*apinetworkingv1.Ingress)

						if err := dm.HandleIngressExists(ingress); err != nil {
							log.Error(err)
						}
					},
					DeleteFunc: func(obj interface{}) {
						ingress := obj.(*apinetworkingv1.Ingress)

						if err := dm.HandleIngressDeleted(ingress); err != nil {
							log.Error(err)
						}
					},
					UpdateFunc: func(old, new interface{}) {
						ingress := new.(*apinetworkingv1.Ingress)

						if err := dm.HandleIngressExists(ingress); err != nil {
							log.Error(err)
						}
					},
				},
			)

			stop := make(chan struct{})
			informerFactory.Start(stop)
			informerFactory.WaitForCacheSync(stop)

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

			<-sig
			stop <- struct{}{}

			return nil
		},
	}

	watchCmd.Flags().StringVarP(&ingressClass, "ingress-class", "i", "", "ingress class to use for public DNS records")
	watchCmd.Flags().StringVarP(&domainName, "domain", "d", "", "domain name for API calls")

	return watchCmd
}
