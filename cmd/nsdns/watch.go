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

			informerFactory := DomainManagerInformerFactory(dm, clientset)

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
