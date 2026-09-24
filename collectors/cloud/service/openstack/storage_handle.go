// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package openstack

import (
	"context"
	"log/slog"
	"strings"
	"time"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/constants"
	"confirmate.io/core/api/ontology"

	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v2/backups"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumetypes"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/v2/pagination"
	"github.com/lmittmann/tint"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handleBlockStorage creates a block storage resource based on the CSC Hub Ontology
func (d *openstackCollector) handleBlockStorage(volume *volumes.Volume) (ontology.IsResource, error) {
	var (
		are    *ontology.AtRestEncryption
		backup []*ontology.Backup
	)

	// Get Name, if exits, otherwise take the ID
	name := volume.Name
	if volume.Name == "" {
		name = volume.ID
	}

	// Get encryption information. Unfortunately, this requires a second lookup of the volume type.
	vType, err := volumetypes.Get(context.Background(), d.clients.blockStorageClient, volume.VolumeType).Extract()
	if err != nil {
		log.Error("error getting volume type information for volume", slog.String("name", volume.Name), tint.Err(err))
	} else {
		enc, err := volumetypes.GetEncryption(context.Background(), d.clients.blockStorageClient, vType.ID).Extract()
		if err != nil {
			log.Error("error getting encryption information for volume", slog.String("name", volume.Name), tint.Err(err))
		} else if enc.EncryptionID != "" {
			are = &ontology.AtRestEncryption{
				Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
					CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
						Enabled:   new(true),
						Algorithm: new(enc.Cipher),
					},
				},
			}
		}
	}

	// Get backup information. OpenStack does not provide a direct way to check if backups are enabled for a
	// volume, so we use the presence of any associated backup as a heuristic.
	err = backups.List(d.clients.blockStorageClient, backups.ListOpts{
		VolumeID: volume.ID,
	}).EachPage(context.Background(), func(_ context.Context, p pagination.Page) (bool, error) {
		backupList, err := backups.ExtractBackups(p)
		if err != nil {
			return false, err
		}

		for _, b := range backupList {
			backup = append(backup, &ontology.Backup{
				StorageId:       new(b.ID),
				RetentionPeriod: durationpb.New(0 * time.Second), // retention period is unlimited
				Enabled:         new(true),
			})

			log.Info("Adding block storage backup", slog.String("name", b.Name))
		}

		return true, nil
	})
	if err != nil {
		log.Error("error listing backups for block storage", slog.String("name", volume.Name), tint.Err(err))
	}

	r := &ontology.BlockStorage{
		Id:           new(volume.ID),
		Name:         new(name),
		Description:  new(volume.Description),
		CreationTime: timestamppb.New(volume.CreatedAt),
		GeoLocation: &ontology.GeoLocation{
			Region: new(d.region),
		},
		ParentId:         new(getParentID(volume)),
		Labels:           map[string]string{}, // Not available
		Raw:              new(collector.Raw(volume)),
		AtRestEncryption: are,
		Backups:          backup,
	}

	log.Info("Adding block storage", slog.String("name", volume.Name))

	return r, nil
}

// handleObjectStorage creates an object storage resource based on the CSC Hub Ontology
func (d *openstackCollector) handleObjectStorage(container *containers.Container) (ontology.IsResource, error) {
	var isPublic bool

	header, err := containers.Get(context.Background(), d.clients.storageClient, container.Name, containers.GetOpts{}).Extract()
	if err != nil {
		log.Error("error extracting container details for container", slog.String("name", container.Name), tint.Err(err))
	} else {
		// A container is public if its "X-Container-Read" ACL grants read access to everyone (".r:*").
		for _, acl := range header.Read {
			if strings.Contains(acl, ".r:*") {
				isPublic = true
				break
			}
		}
	}

	r := &ontology.ObjectStorage{
		Id:           new(container.Name),
		Name:         new(container.Name),
		ParentId:     new(d.clients.storageClient.Endpoint), // parent is the object storage service stored in the endpoint
		PublicAccess: new(isPublic),
	}

	log.Info("Adding object storage", slog.String("name", container.Name))

	return r, nil
}

// handleObjectStorageService creates an object storage service resource based on the CSC Hub Ontology
func (d *openstackCollector) handleObjectStorageService() (ontology.IsResource, error) {
	var te *ontology.TransportEncryption

	if strings.HasPrefix(d.clients.storageClient.Endpoint, "https://") {
		te = &ontology.TransportEncryption{
			Enabled:  new(true),
			Enforced: new(true),
			Protocol: new(constants.TLS),
		}
	}

	r := &ontology.ObjectStorageService{
		Id:   new(d.clients.storageClient.Endpoint),
		Name: new("Swift Object Storage Service"),
		GeoLocation: &ontology.GeoLocation{
			Region: new(d.region),
		},
		ParentId: new(d.project.projectID),
		HttpEndpoint: &ontology.HttpEndpoint{
			Url:                 new(d.clients.storageClient.Endpoint),
			TransportEncryption: te,
		},
	}

	return r, nil
}
