// Copyright 2016-2025 Fraunhofer AISEC
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

package orchestrator

import (
	"context"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

func TestService_CreateCatalog(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.CreateCatalogRequest
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.CreateCatalogRequest{
					Catalog: &orchestrator.Catalog{
						Id:          "catalog-1",
						Name:        "Test Catalog",
						Description: "A test catalog",
					},
				},
				wantErr: false,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(tt.req)
			)

			res, err := svc.CreateCatalog(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Catalog.Id, res.Msg.Id)
		})
	}
}

func TestService_GetCatalog(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "catalog-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Catalog{
						Id:          "catalog-1",
						Name:        "Test Catalog",
						Description: "A test catalog",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name:    "not found",
				id:      "non-existent",
				setup:   func(svc *service) {},
				wantErr: true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.GetCatalogRequest{
					CatalogId: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.GetCatalog(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.id, res.Msg.Id)
		})
	}
}

func TestService_ListCatalogs(t *testing.T) {
	var (
		tests = []struct {
			name      string
			setup     func(*service)
			wantCount int
		}{
			{
				name: "list all",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Catalog{
						Id:          "catalog-1",
						Name:        "Catalog 1",
						Description: "First catalog",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.Catalog{
						Id:          "catalog-2",
						Name:        "Catalog 2",
						Description: "Second catalog",
					})
					assert.NoError(t, err)
				},
				wantCount: 2,
			},
			{
				name:      "empty list",
				setup:     func(svc *service) {},
				wantCount: 0,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.ListCatalogsRequest{})
			)

			tt.setup(svc)

			res, err := svc.ListCatalogs(context.Background(), req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantCount, len(res.Msg.Catalogs))
		})
	}
}

func TestService_UpdateCatalog(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.UpdateCatalogRequest
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.UpdateCatalogRequest{
					Catalog: &orchestrator.Catalog{
						Id:          "catalog-1",
						Name:        "Updated Catalog",
						Description: "Updated description",
					},
				},
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Catalog{
						Id:          "catalog-1",
						Name:        "Test Catalog",
						Description: "Original description",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name: "not found",
				req: &orchestrator.UpdateCatalogRequest{
					Catalog: &orchestrator.Catalog{
						Id:          "non-existent",
						Name:        "Updated Catalog",
						Description: "Updated description",
					},
				},
				setup:   func(svc *service) {},
				wantErr: true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(tt.req)
			)

			tt.setup(svc)

			res, err := svc.UpdateCatalog(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Catalog.Name, res.Msg.Name)
		})
	}
}

func TestService_RemoveCatalog(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "catalog-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Catalog{
						Id:          "catalog-1",
						Name:        "Test Catalog",
						Description: "A test catalog",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.RemoveCatalogRequest{
					CatalogId: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.RemoveCatalog(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}

func TestService_GetCategory(t *testing.T) {
	var (
		tests = []struct {
			name         string
			catalogId    string
			categoryName string
			setup        func(*service)
			wantErr      bool
		}{
			{
				name:         "happy path",
				catalogId:    "catalog-1",
				categoryName: "category-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Category{
						Name:      "category-1",
						CatalogId: "catalog-1",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name:         "not found",
				catalogId:    "catalog-1",
				categoryName: "non-existent",
				setup:        func(svc *service) {},
				wantErr:      true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.GetCategoryRequest{
					CatalogId:    tt.catalogId,
					CategoryName: tt.categoryName,
				})
			)

			tt.setup(svc)

			res, err := svc.GetCategory(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.categoryName, res.Msg.Name)
		})
	}
}

func TestService_ListControls(t *testing.T) {
	var (
		tests = []struct {
			name         string
			catalogId    string
			categoryName string
			setup        func(*service)
			wantCount    int
		}{
			{
				name:      "list all",
				catalogId: "",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Control{
						Id:               "control-1",
						CategoryName:     "category-1",
						CategoryCatalogId: "catalog-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.Control{
						Id:               "control-2",
						CategoryName:     "category-2",
						CategoryCatalogId: "catalog-2",
					})
					assert.NoError(t, err)
				},
				wantCount: 2,
			},
			{
				name:      "filter by catalog",
				catalogId: "catalog-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Control{
						Id:               "control-1",
						CategoryName:     "category-1",
						CategoryCatalogId: "catalog-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.Control{
						Id:               "control-2",
						CategoryName:     "category-2",
						CategoryCatalogId: "catalog-2",
					})
					assert.NoError(t, err)
				},
				wantCount: 1,
			},
			{
				name:         "filter by category",
				categoryName: "category-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Control{
						Id:               "control-1",
						CategoryName:     "category-1",
						CategoryCatalogId: "catalog-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&orchestrator.Control{
						Id:               "control-2",
						CategoryName:     "category-2",
						CategoryCatalogId: "catalog-1",
					})
					assert.NoError(t, err)
				},
				wantCount: 1,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.ListControlsRequest{
					CatalogId:    tt.catalogId,
					CategoryName: tt.categoryName,
				})
			)

			tt.setup(svc)

			res, err := svc.ListControls(context.Background(), req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantCount, len(res.Msg.Controls))
		})
	}
}

func TestService_GetControl(t *testing.T) {
	var (
		tests = []struct {
			name         string
			controlId    string
			categoryName string
			catalogId    string
			setup        func(*service)
			wantErr      bool
		}{
			{
				name:         "happy path",
				controlId:    "control-1",
				categoryName: "category-1",
				catalogId:    "catalog-1",
				setup: func(svc *service) {
					err := svc.db.Create(&orchestrator.Control{
						Id:               "control-1",
						CategoryName:     "category-1",
						CategoryCatalogId: "catalog-1",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name:         "not found",
				controlId:    "non-existent",
				categoryName: "category-1",
				catalogId:    "catalog-1",
				setup:        func(svc *service) {},
				wantErr:      true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.GetControlRequest{
					ControlId:    tt.controlId,
					CategoryName: tt.categoryName,
					CatalogId:    tt.catalogId,
				})
			)

			tt.setup(svc)

			res, err := svc.GetControl(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.controlId, res.Msg.Id)
		})
	}
}
