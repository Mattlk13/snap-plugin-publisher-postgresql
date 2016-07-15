// +build integration

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package postgresql

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPostgresPublish(t *testing.T) {
	config := make(map[string]ctypes.ConfigValue)

	Convey("Snap Plugin PostgreSQL integration testing with PostgreSQL", t, func() {
		var buf bytes.Buffer

		config["hostname"] = ctypes.ConfigValueStr{Value: os.Getenv("SNAP_POSTGRESQL_HOST")}
		config["port"] = ctypes.ConfigValueInt{Value: 5432}
		config["username"] = ctypes.ConfigValueStr{Value: "postgres"}
		config["password"] = ctypes.ConfigValueStr{Value: ""}
		config["database"] = ctypes.ConfigValueStr{Value: "snap_test"}
		config["table_name"] = ctypes.ConfigValueStr{Value: "info"}

		ip := NewPostgreSQLPublisher()
		cp, _ := ip.GetConfigPolicy()
		cfg, _ := cp.Get([]string{""}).Process(config)

		Convey("Publish integer metric (int)", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), nil, "", 99),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish integer metric (int64)", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), nil, "", int64(99)),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish unsigned integer metric (uint)", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), nil, "", uint(99)),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish unsigned integer metric (uint64)", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), nil, "", uint64(99)),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish float metric", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("bar"), time.Now(), nil, "", 3.141),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish string metric", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("qux"), time.Now(), nil, "", "bar"),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish boolean metric", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("baz"), time.Now(), nil, "", true),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish multiple metrics", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), nil, "", 101),
				*plugin.NewMetricType(core.NewNamespace("bar"), time.Now(), nil, "", 5.789),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

	})
}
