package postgis_test

import (
	"context"
	"testing"

	"github.com/go-spatial/tegola"
	"github.com/go-spatial/tegola/internal/ttools"
	"github.com/go-spatial/tegola/provider"
	"github.com/go-spatial/tegola/provider/postgis"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDBConfig(t *testing.T) {
	uri := ttools.GetEnvDefault("PGURI", "postgres://postgres:postgres@localhost:5432/tegola")

	type tcase struct {
		opts                          *postgis.DBConfigOptions
		expApplicationName            string
		expDefaultTransactionReadOnly string
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			dbconfig, err := postgis.BuildDBConfig(
				tc.opts)
			if err != nil {
				t.Errorf("unable to build config: %v", err)
			}

			applicationName := dbconfig.ConnConfig.RuntimeParams["application_name"]
			if applicationName != tc.expApplicationName {
				t.Errorf("expected application name: %s, got: %s", tc.expApplicationName, applicationName)
			}

			defaultTransactionReadOnly := dbconfig.ConnConfig.RuntimeParams["default_transaction_read_only"]
			if defaultTransactionReadOnly != tc.expDefaultTransactionReadOnly {
				t.Errorf("expected transaction read only: %s, got: %s", tc.expDefaultTransactionReadOnly, defaultTransactionReadOnly)
			}
		}
	}
	tests := map[string]tcase{
		"1": {
			opts: &postgis.DBConfigOptions{
				Uri:                        uri,
				ApplicationName:            "tegola",
				DefaultTransactionReadOnly: "TRUE",
			},
			expApplicationName:            "tegola",
			expDefaultTransactionReadOnly: "TRUE",
		},
		"2": {
			opts: &postgis.DBConfigOptions{
				Uri:                        uri,
				ApplicationName:            "aloget",
				DefaultTransactionReadOnly: "OFF",
			},
			expApplicationName:            "aloget",
			expDefaultTransactionReadOnly: "",
		},
		"3": {
			opts: &postgis.DBConfigOptions{
				Uri:                        uri,
				ApplicationName:            "tegola",
				DefaultTransactionReadOnly: "FALSE",
			},
			expApplicationName:            "tegola",
			expDefaultTransactionReadOnly: "FALSE",
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestTLSConfig(t *testing.T) {
	uri := "postgres://testuser:testpassword@testhost:5432/testdb"

	testConnConfig, err := postgis.BuildDBConfig(
		&postgis.DBConfigOptions{
			Uri:                        uri,
			DefaultTransactionReadOnly: "TRUE",
			ApplicationName:            "tegola",
		})
	if err != nil {
		t.Fatalf("unable to build db config: %v", err)
	}

	type tcase struct {
		sslMode     string
		sslKey      string
		sslCert     string
		sslRootCert string
		testFunc    func(config *pgxpool.Config)
		shouldError bool
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			err := postgis.ConfigTLS(tc.sslMode, tc.sslKey, tc.sslCert, tc.sslRootCert, testConnConfig)
			if !tc.shouldError && err != nil {
				t.Errorf("unable to create a new provider: %v", err)
				return
			} else if tc.shouldError && err == nil {
				t.Errorf("Error expected but got no error")
				return
			}

			tc.testFunc(testConnConfig)
		}
	}

	tests := map[string]tcase{
		"1": {
			sslMode:     "",
			sslKey:      "",
			sslCert:     "",
			sslRootCert: "",
			shouldError: true,
			testFunc: func(config *pgxpool.Config) {
			},
		},
		"2": {
			sslMode:     "disable",
			sslKey:      "",
			sslCert:     "",
			sslRootCert: "",
			shouldError: false,
			testFunc: func(config *pgxpool.Config) {
				if config.ConnConfig.TLSConfig != nil {
					t.Errorf("When using disable ssl mode; UseFallbackTLS, expected nil got %v", testConnConfig.ConnConfig.TLSConfig)
				}
			},
		},
		"3": {
			sslMode:     "allow",
			sslKey:      "",
			sslCert:     "",
			sslRootCert: "",
			shouldError: false,
			testFunc: func(config *pgxpool.Config) {
				if config.ConnConfig.TLSConfig.InsecureSkipVerify == false {
					t.Error("When using allow ssl mode; UseFallbackTLS.InsecureSkipVerify, expected true got false")
				}
			},
		},
		"4": {
			sslMode:     "prefer",
			sslKey:      "",
			sslCert:     "",
			sslRootCert: "",
			shouldError: false,
			testFunc: func(config *pgxpool.Config) {
				if config.ConnConfig.TLSConfig == nil {
					t.Error("When using prefer ssl mode; TLSConfig, expected not nil got nil")
				}

				if config.ConnConfig.TLSConfig != nil && config.ConnConfig.TLSConfig.InsecureSkipVerify == false {
					t.Error("When using prefer ssl mode; TLSConfig.InsecureSkipVerify, expected true got false")
				}
			},
		},
		"5": {
			sslMode:     "require",
			sslKey:      "",
			sslCert:     "",
			sslRootCert: "",
			shouldError: false,
			testFunc: func(config *pgxpool.Config) {
				if config.ConnConfig.TLSConfig == nil {
					t.Error("When using prefer ssl mode; TLSConfig, expected not nil got nil")
				}

				if config.ConnConfig.TLSConfig != nil && config.ConnConfig.TLSConfig.InsecureSkipVerify == false {
					t.Error("When using prefer ssl mode; TLSConfig.InsecureSkipVerify, expected true got false")
				}
			},
		},
		"6": {
			sslMode:     "verify-ca",
			sslKey:      "",
			sslCert:     "",
			sslRootCert: "",
			shouldError: false,
			testFunc: func(config *pgxpool.Config) {
				if config.ConnConfig.TLSConfig == nil {
					t.Error("When using prefer ssl mode; TLSConfig, expected not nil got nil")
				}

				if config.ConnConfig.TLSConfig != nil && config.ConnConfig.TLSConfig.ServerName != testConnConfig.ConnConfig.Host {
					t.Errorf("When using prefer ssl mode; TLSConfig.ServerName, expected %s got %s", testConnConfig.ConnConfig.Host, config.ConnConfig.TLSConfig.ServerName)
				}
			},
		},
		"7": {
			sslMode:     "verify-full",
			sslKey:      "",
			sslCert:     "",
			sslRootCert: "",
			shouldError: false,
			testFunc: func(config *pgxpool.Config) {
				if config.ConnConfig.TLSConfig == nil {
					t.Error("When using prefer ssl mode; TLSConfig, expected not nil got nil")
				}

				if config.ConnConfig.TLSConfig != nil && config.ConnConfig.TLSConfig.ServerName != testConnConfig.ConnConfig.Host {
					t.Errorf("When using prefer ssl mode; TLSConfig.ServerName, expected %s got %s", testConnConfig.ConnConfig.Host, config.ConnConfig.TLSConfig.ServerName)
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestNewTileProvider(t *testing.T) {
	ttools.ShouldSkip(t, postgis.TESTENV)

	fn := func(tc postgis.TCConfig) func(t *testing.T) {
		return func(t *testing.T) {
			config := tc.Config(postgis.DefaultEnvConfig)
			config[postgis.ConfigKeyName] = "provider_name"
			_, err := postgis.NewTileProvider(config, nil)
			if err != nil {
				t.Errorf("unable to create a new provider. err: %v", err)
				return
			}
		}
	}

	tests := map[string]postgis.TCConfig{
		"1": {
			LayerConfig: []map[string]interface{}{
				{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeyTablename: "ne_10m_land_scale_rank",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestTileFeatures(t *testing.T) {
	ttools.ShouldSkip(t, postgis.TESTENV)

	type tcase struct {
		postgis.TCConfig
		tile                 provider.Tile
		expectedErr          error
		expectedFeatureCount int
		expectedTags         []string
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			config := tc.Config(postgis.DefaultEnvConfig)
			config[postgis.ConfigKeyName] = "provider_name"
			p, err := postgis.NewTileProvider(config, nil)
			if err != nil {
				t.Errorf("unexpected error; unable to create a new provider, expected: nil Got %v", err)
				return
			}

			layerName := tc.LayerConfig[0][postgis.ConfigKeyLayerName].(string)

			var featureCount int
			err = p.TileFeatures(context.Background(), layerName, tc.tile, nil, func(f *provider.Feature) error {
				// only verify tags on first feature
				if featureCount == 0 {
					for _, tag := range tc.expectedTags {
						if _, ok := f.Tags[tag]; !ok {
							t.Errorf("expected tag %v in %v", tag, f.Tags)
							return nil
						}
					}
				}

				featureCount++

				return nil
			})
			if err != tc.expectedErr {
				t.Errorf("expected err (%v) got err (%v)", tc.expectedErr, err)
				return
			}

			if featureCount != tc.expectedFeatureCount {
				t.Errorf("feature count, expected %v got %v", tc.expectedFeatureCount, featureCount)
				return
			}
		}
	}

	tests := map[string]tcase{
		"tablename query": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeyTablename: "ne_10m_land_scale_rank",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 4032,
			expectedTags:         []string{"scalerank", "featurecla"},
		},
		"tablename query with fields": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeyTablename: "ne_10m_land_scale_rank",
					postgis.ConfigKeyFields:    []string{"scalerank"},
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 4032,
			expectedTags:         []string{"scalerank"},
		},
		"tablename query with fields and id as field": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName:   "land",
					postgis.ConfigKeyTablename:   "ne_10m_land_scale_rank",
					postgis.ConfigKeyGeomIDField: "gid",
					postgis.ConfigKeyFields:      []string{"gid", "scalerank"},
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 4032,
			expectedTags:         []string{"gid", "scalerank"},
		},
		"SQL sub-query": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       "(SELECT gid, geom, featurecla FROM ne_10m_land_scale_rank LIMIT 100) AS sub",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 100,
			expectedTags:         []string{"featurecla"},
		},
		"SQL sub-query multi line": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL: ` (
					SELECT gid, geom, featurecla FROM ne_10m_land_scale_rank LIMIT 100
				) AS sub`,
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 100,
			expectedTags:         []string{"featurecla"},
		},
		"SQL sub-query and tablename": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       "(SELECT gid, geom, featurecla FROM ne_10m_land_scale_rank LIMIT 100) AS sub",
					postgis.ConfigKeyTablename: "not_good_name",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 100,
			expectedTags:         []string{"featurecla"},
		},
		"SQL sub-query space after prens": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       "(  SELECT gid, geom, featurecla FROM ne_10m_land_scale_rank LIMIT 100) AS sub",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 100,
			expectedTags:         []string{"featurecla"},
		},
		"SQL sub-query space before prens": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       "   (SELECT gid, geom, featurecla FROM ne_10m_land_scale_rank LIMIT 100) AS sub",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 100,
			expectedTags:         []string{"featurecla"},
		},
		"SQL sub-query with comments": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       " -- this is a comment\n-- accross multiple lines\n (SELECT gid, geom, scalerank FROM ne_10m_land_scale_rank LIMIT 100) AS sub -- another comment at the end",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 100,
			expectedTags:         []string{"scalerank"},
		},
		"SQL sub-query with *": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       "(SELECT * FROM ne_10m_land_scale_rank LIMIT 100) AS sub",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 100,
			expectedTags:         []string{"scalerank", "featurecla"},
		},
		"SQL sub-query with * and fields": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       "(SELECT * FROM ne_10m_land_scale_rank LIMIT 100) AS sub",
					postgis.ConfigKeyFields:    []string{"scalerank"},
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 100,
			expectedTags:         []string{"scalerank"},
		},
		"SQL with !ZOOM!": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       "SELECT gid, ST_AsBinary(geom) AS geom FROM ne_10m_land_scale_rank WHERE scalerank=!ZOOM! AND geom && !BBOX!",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 98,
		},
		"SQL sub-query with token in SELECT": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeyGeomType:  "polygon", // required to disable SQL inspection
					postgis.ConfigKeySQL:       "(SELECT gid, geom, !ZOOM! * 2 AS doublezoom FROM ne_10m_land_scale_rank WHERE scalerank = !ZOOM! AND geom && !BBOX!) AS sub",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 98,
			expectedTags:         []string{"doublezoom"},
		},
		"SQL sub-query with fields": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       "(SELECT gid, geom, 1 AS a, '2' AS b, 3 AS c FROM ne_10m_land_scale_rank WHERE scalerank = !ZOOM! AND geom && !BBOX!) AS sub",
					postgis.ConfigKeyFields:    []string{"gid", "a", "b"},
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 98,
			expectedTags:         []string{"a", "b"},
			// expectedTags:         []string{"gid", "a", "b"}, TODO #383
		},
		"SQL with comments": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "land",
					postgis.ConfigKeySQL:       " -- this is a comment\n -- accross multiple lines \n \tSELECT gid, -- gid \nST_AsBinary(geom) AS geom -- geom \n FROM ne_10m_land_scale_rank WHERE scalerank=!ZOOM! AND geom && !BBOX! -- comment at the end",
				}},
			},
			tile:                 provider.NewTile(1, 1, 1, 64, tegola.WebMercator),
			expectedFeatureCount: 98,
		},
		"decode numeric(x,x) types": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName:   "buildings",
					postgis.ConfigKeyGeomIDField: "osm_id",
					postgis.ConfigKeyGeomField:   "geometry",
					postgis.ConfigKeySQL:         "SELECT ST_AsBinary(geometry) AS geometry, osm_id, name, nullif(as_numeric(height),-1) AS height, type FROM osm_buildings_test WHERE geometry && !BBOX!",
				}},
			},
			tile:                 provider.NewTile(16, 11241, 26168, 64, tegola.WebMercator),
			expectedFeatureCount: 101,
			expectedTags:         []string{"name", "type"}, // height can be null and therefore missing from the tags
		},
		"gracefully handle 3d point": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName:   "three_d_points",
					postgis.ConfigKeyGeomIDField: "id",
					postgis.ConfigKeyGeomField:   "geom",
					postgis.ConfigKeySQL:         "SELECT ST_AsBinary(geom) AS geom, id FROM three_d_test WHERE geom && !BBOX!",
				}},
			},
			tile:                 provider.NewTile(0, 0, 0, 64, tegola.WebMercator),
			expectedFeatureCount: 0,
		},
		"gracefully handle null geometry": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName:   "null_geom",
					postgis.ConfigKeyGeomIDField: "id",
					postgis.ConfigKeyGeomField:   "geometry",
					// this SQL is a workaround the normal !BBOX! WHERE clause. we're simulating a null geometry lookup in the table and don't want to filter by bounding box
					postgis.ConfigKeySQL: "SELECT id, ST_AsBinary(geometry) AS geometry, !BBOX! AS bbox FROM null_geom_test",
				}},
			},
			tile:                 provider.NewTile(16, 11241, 26168, 64, tegola.WebMercator),
			expectedFeatureCount: 1,
			expectedTags:         []string{"bbox"},
		},
		"missing geom field name": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "missing_geom_field_name",
					postgis.ConfigKeyGeomField: "geom",
					// this SQL is a workaround the normal !BBOX! token check. We don't care about the bounding
					// box query, but rather simulating the missing geom column to trigger the error we're testing for.
					postgis.ConfigKeySQL: "SELECT ST_AsBinary(geom), !BBOX! AS bbox FROM three_d_test",
				}},
			},
			tile: provider.NewTile(16, 11241, 26168, 64, tegola.WebMercator),
			expectedErr: postgis.ErrGeomFieldNotFound{
				GeomFieldName: "geom",
				LayerName:     "missing_geom_field_name",
			},
		},
		"empty geometry collection": {
			TCConfig: postgis.TCConfig{
				LayerConfig: []map[string]interface{}{{
					postgis.ConfigKeyLayerName: "empty_geometry_collection",
					postgis.ConfigKeyGeomField: "geom",
					postgis.ConfigKeyGeomType:  "polygon", // bypass the geometry type sniff on init
					postgis.ConfigKeySQL:       "SELECT ST_AsBinary(ST_GeomFromText('GEOMETRYCOLLECTION EMPTY')) AS geom, !BBOX! AS bbox",
				}},
			},
			tile:                 provider.NewTile(16, 11241, 26168, 64, tegola.WebMercator),
			expectedFeatureCount: 1,
			expectedTags:         []string{},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}
