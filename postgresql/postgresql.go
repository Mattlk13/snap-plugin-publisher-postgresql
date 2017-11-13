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
	"fmt"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"database/sql"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
	// Import of postgresql library
	_ "github.com/lib/pq"
)

const (
	name         = "postgresql"
	version      = 9
	pluginType   = plugin.PublisherPluginType
	tableColumns = "(id SERIAL PRIMARY KEY, time_posted timestamp with time zone, key_column VARCHAR(200), value_column VARCHAR(200))"
	timeFormat   = time.RFC3339
)

// PostgreSQLPublisher struct
type PostgreSQLPublisher struct {
}

// NewPostgreSQLPublisher return new PostgreSQL instance
func NewPostgreSQLPublisher() *PostgreSQLPublisher {
	return &PostgreSQLPublisher{}
}

// Publish sends data to PostgreSQL server
func (s *PostgreSQLPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := log.New()
	logger.Println("Publishing started")
	var metrics []plugin.MetricType

	switch contentType {
	case plugin.SnapGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			logger.Printf("Error decoding: error=%v content=%v", err, content)
			return err
		}
	default:
		logger.Printf("Error unknown content type '%v'", contentType)
		return fmt.Errorf("Unknown content type '%s'", contentType)
	}

	logger.Printf("publishing %v to %v", metrics, config)

	tableName := config["table_name"].(ctypes.ConfigValueStr).Value

	// Open connection and ping to make sure it works
	db, err := getPostgreSQLConn(config)
	if err != nil {
		logger.Printf("Error: %v", err)
		return err
	}

	defer db.Close()

	nowTime := time.Now().Format(timeFormat)
	var key, value string
	for _, m := range metrics {
		key = sliceToNamespace(m.Namespace().Strings())
		value, err = interfaceToString(m.Data())
		if err == nil {
			query := fmt.Sprintf("INSERT INTO %s (id, time_posted, key_column, value_column) VALUES (DEFAULT, '%s', '%s', '%s')", tableName, nowTime, key, value)
			_, err := db.Exec(query)
			if err != nil {
				errMsg := fmt.Sprintf("pq: relation \"%s\" does not exist", tableName)
				if err.Error() == errMsg {
					_, err = createTable(db, tableName)
					if err != nil {
						logger.Printf("Error: %v", err)
						return err
					}

				}
				logger.Printf("Error: %v", err)
				return err
			}
		} else {
			logger.Printf("Error: %v", err)
			return err
		}
	}
	return nil
}

// Meta returns plugin meta data info
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

func getPostgreSQLConn(config map[string]ctypes.ConfigValue) (*sql.DB, error) {
	logger := log.New()
	hostName := config["hostname"].(ctypes.ConfigValueStr).Value
	port := config["port"].(ctypes.ConfigValueInt).Value
	username := config["username"].(ctypes.ConfigValueStr).Value
	password := config["password"].(ctypes.ConfigValueStr).Value
	database := config["database"].(ctypes.ConfigValueStr).Value
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", hostName, port, username, password, database)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		logger.Printf("Error: %v", err)
		return db, err
	}
	err = db.Ping()
	if err != nil {
		logger.Printf("Error: %v", err)
		return db, err
	}
	return db, err
}

func createTable(db *sql.DB, tableName string) (bool, error) {
	logger := log.New()
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s", tableName, tableColumns)
	_, err := db.Exec(query)
	if err != nil {
		logger.Printf("Error: %v", err)
		return false, err
	}
	query = fmt.Sprintf("CREATE INDEX key_index on %s (key_column)", tableName)
	_, err = db.Exec(query)
	if err != nil {
		logger.Printf("Error: %v", err)
		return false, err
	}
	return true, err
}

// GetConfigPolicy returns a config policy
func (s *PostgreSQLPublisher) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	username, err := cpolicy.NewStringRule("username", true)
	handleErr(err)
	username.Description = "Username to login to the PostgreSQL server"

	password, err := cpolicy.NewStringRule("password", true)
	handleErr(err)
	password.Description = "Password to login to the PostgreSQL server"

	database, err := cpolicy.NewStringRule("database", true)
	handleErr(err)
	database.Description = "The postgresql database that data will be pushed to"

	tableName, err := cpolicy.NewStringRule("table_name", true)
	handleErr(err)
	tableName.Description = "The postgresql table within the database where information will be stored"

	hostName, err := cpolicy.NewStringRule("hostname", true, "localhost")
	handleErr(err)
	tableName.Description = "The postgresql server ip or domain name"

	port, err := cpolicy.NewIntegerRule("port", true, 5432)
	handleErr(err)
	port.Description = "The postgresql server port number"

	config.Add(username, password, database, tableName, hostName, port)

	cp.Add([]string{""}, config)
	return cp, nil

}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}

func sliceToString(slice []string) string {
	return strings.Join(slice, ", ")
}

func sliceToNamespace(slice []string) string {
	return strings.Join(slice, ".")
}

func interfaceToString(face interface{}) (string, error) {
	var (
		ret string
		err error
	)

	switch face.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		string:

		ret = fmt.Sprintf("%v", face)

	case []uint, []uint8, []uint16, []uint32, []uint64,
		[]int, []int8, []int16, []int32, []int64,
		[]float32, []float64:
		ret = strings.Trim(strings.Replace(fmt.Sprintf("%v", face), " ", ", ", -1), "[]")
	case []string:
		//special case for slice of strings to deal with spaces and `[]` in elements of slice
		ret = strings.Join(face.([]string), ", ")
	case bool:
		if face.(bool) {
			ret = "1"
		} else {
			ret = "0"
		}
	default:
		err = fmt.Errorf("Unsupported type %v (currently supported data types: bool, "+
			"int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string"+
			"[]uint, []uint8, []uint16, []uint32, []uint64, []int, []int8, []int16, []int32, []int64, []float32, []float64, []string)", reflect.TypeOf(face))

	}
	return ret, err
}
