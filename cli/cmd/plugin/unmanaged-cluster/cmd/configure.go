// Copyright 2020-2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/community-edition/cli/cmd/plugin/unmanaged-cluster/config"
	logger "github.com/vmware-tanzu/community-edition/cli/cmd/plugin/unmanaged-cluster/log"
)

const configureDesc = `
Generate a configuration file that can be used when running:
tanzu unmanaged-cluster create -f <config-file-name>.yaml. Configure generates
a config file injected with default values. When flags are specified
(e.g. --cni) the flag value is respected in the overridden config.
`

// ConfigureCmd creates an unmanaged workload cluster.
var ConfigureCmd = &cobra.Command{
	Use:     "configure <cluster name>",
	Aliases: []string{"config", "conf"},
	Short:   "Generate a config file to be used in cluster creation",
	Long:    configureDesc,
	RunE:    configure,
}

func init() {
	ConfigureCmd.Flags().StringVarP(&co.clusterConfigFile, "config", "f", "", "Configuration file for unmanaged cluster creation")
	ConfigureCmd.Flags().StringVar(&co.infrastructureProvider, "provider", "", "The infrastructure provider to use for cluster creation. Default is 'kind'")
	ConfigureCmd.Flags().StringVarP(&co.tkrLocation, "tkr", "t", "", "The URL to the image or path to local file containing a Tanzu Kubernetes release")
	ConfigureCmd.Flags().StringVarP(&co.cni, "cni", "c", "", "The CNI to deploy. Default is 'calico'")
	ConfigureCmd.Flags().StringVar(&co.podcidr, "pod-cidr", "", "The CIDR to use for Pod IP addresses. Default and format is '10.244.0.0/16'")
	ConfigureCmd.Flags().StringVar(&co.servicecidr, "service-cidr", "", "The CIDR to use for Service IP addresses. Default and format is '10.96.0.0/16'")
	ConfigureCmd.Flags().Bool("tty-disable", false, "Disable log stylization and emojis")
	ConfigureCmd.Flags().StringSliceVar(&co.additionalRepo, "additional-repo", []string{}, "Addresses for additional package repositories to install")
	ConfigureCmd.Flags().StringSliceVar(&co.profile, "profile", []string{}, "(experimental) A profile to install. May be specified multiple times. Profile mappings supported - profile-name:profile-version:profile-config-file. profile-name should be the fully qualified package name or a prefix to a package name found in an installed package repository. profile-version is optional and resolves to the latest semantic versioned package if not specified or `latest` is entered. package-config-file is optional and should be the path to a values yaml file in order to configure the package.")
}

func configure(cmd *cobra.Command, args []string) error {
	var clusterName string

	// validate a cluster name was passed
	if len(args) < 1 {
		return fmt.Errorf("cluster name not specified")
	} else if len(args) == 1 {
		clusterName = args[0]
	}

	log := logger.NewLogger(TtySetting(cmd.Flags()), 0)

	profiles, err := config.ParseProfileMappings(co.profile)
	if err != nil {
		log.Error(err.Error())
	}

	// Determine our configuration to use
	configArgs := map[string]interface{}{
		config.ClusterConfigFile:      co.clusterConfigFile,
		config.ClusterName:            clusterName,
		config.Provider:               co.infrastructureProvider,
		config.TKRLocation:            co.tkrLocation,
		config.Cni:                    co.cni,
		config.PodCIDR:                co.podcidr,
		config.ServiceCIDR:            co.servicecidr,
		config.AdditionalPackageRepos: co.additionalRepo,
		config.Profiles:               profiles,
	}

	scConfig, err := config.InitializeConfiguration(configArgs)
	if err != nil {
		log.Errorf("Failed to initialize configuration. Error: %s\n", err.Error())
		return nil
	}
	fileName := fmt.Sprintf("%s.yaml", clusterName)

	err = config.RenderConfigToFile(fileName, scConfig)
	if err != nil {
		log.Errorf("Failed to write configuration file: %s\n", err.Error())
		return nil
	}
	log.Infof("Wrote configuration file to: %s\n", fileName)

	return nil
}
