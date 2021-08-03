package dirfilestore_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ohler55/ojg/jp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/thingdir/pkg/dirstore"
	"github.com/wostzone/thingdir/pkg/dirstore/dirfilestore"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/vocab"
)

func makeFileStore() *dirfilestore.DirFileStore {
	filename := "/tmp/test-dirfilestore.json"
	store := dirfilestore.NewDirFileStore(filename)
	return store
}

// func readStoredFile() ([]byte, error) {
// 	filename := "/tmp/test-dirfilestore.json"
// 	data, err := ioutil.ReadFile(filename)
// 	return data, err
// }

// Generic directory store testcases
func TestFileStoreStartStop(t *testing.T) {
	fileStore := makeFileStore()
	dirstore.DirStoreStartStop(t, fileStore)
}

func TestFileStoreWrite(t *testing.T) {
	fileStore := makeFileStore()
	dirstore.DirStoreCrud(t, fileStore)
}

func TestQuery(t *testing.T) {
	id1 := "thing1"
	td1 := td.CreateTD(id1, vocab.DeviceTypeSensor)
	td.AddTDProperty(td1, "title", "The sensor")
	id2 := "thing2"
	td2 := td.CreateTD(id2, vocab.DeviceTypeSensor)
	td.AddTDProperty(td2, "title", "The switch")
	fileStore := makeFileStore()
	fileStore.Open()
	// dirstore.DirStoreCrud(t, fileStore)
	tdd := map[string]interface{}(td1)
	fileStore.Replace(id1, tdd)
	tdd = map[string]interface{}(td2)
	fileStore.Replace(id2, tdd)

	t1 := time.Now()
	var i int
	for i = 0; i < 1; i++ {

		// regular filter
		res, err := fileStore.Query(`$[?(@.id=="thing1")]`, 0, 1)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// regular nested filter comparison
		res, err = fileStore.Query(`$[?(@.properties.title=="The sensor")]`, 0, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with nested notation
		res, err = fileStore.Query(`$.*[?(@.title=="The sensor")]`, 0, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with bracket notation
		res, err = fileStore.Query(`$[?(@["id"]=="thing1")]`, 0, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with bracket notation and current object literal (for search @type)
		// only supported by ohler55/ojg
		res, err = fileStore.Query(`$[?(@['@type']=="sensor")]`, 0, 1)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	}
	d1 := time.Since(t1)
	logrus.Infof("TestQuery, %d runs: %d msec", i, d1.Milliseconds())
	// logrus.Infof("Query Results:\n%v", res)
	fileStore.Close()
}

func TestQueryBracketNotationA(t *testing.T) {
	store := make(map[string]interface{})

	jsondoc := `{
		"thing1": {
			"id": "thing1",
			"type": "type1",
			"@type": "sensor",
			"properties": {
				"title": "title1"
			}
		},
		"thing2": {
			"id": "thing2",
			"type": "type2",
			"@type": "sensor",
			"properties": {
				"title": "title1"
			}
		}
	}`
	query1 := `$[?(@['type']=="type1")]`
	query2 := `$[?(@['@type']=="sensor")]`

	err := json.Unmarshal([]byte(jsondoc), &store)
	assert.NoError(t, err)

	jpExpr, err := jp.ParseString(query1)
	assert.NoError(t, err)
	result := jpExpr.Get(store)
	assert.NotEmpty(t, result)

	jpExpr, err = jp.ParseString(query2)
	assert.NoError(t, err)
	result = jpExpr.Get(store)
	assert.NotEmpty(t, result)
}

func TestQueryBracketNotationB(t *testing.T) {
	queryString := "$[?(@['@type']==\"sensor\")]"
	id1 := "thing1"
	td1 := td.CreateTD(id1, vocab.DeviceTypeSensor)
	titleProp := td.CreateProperty("Title", "Sensor title", vocab.PropertyTypeAttr)
	td.AddTDProperty(td1, "title", titleProp)
	valueProp := td.CreateProperty("value", "Sensor value", vocab.PropertyTypeSensor)
	td.AddTDProperty(td1, "title", valueProp)

	id2 := "thing2"
	td2 := make(map[string]interface{})
	td2["id"] = "thing2"
	td2["type"] = "type2"
	td2[vocab.WoTAtType] = "sensor"
	td2["actions"] = make(map[string]interface{})
	td2["properties"] = make(map[string]interface{})
	td.AddTDProperty(td2, "title", "The switch")

	fileStore := makeFileStore()
	fileStore.Open()

	// replace will add if it doesn't exist
	fileStore.Replace(id1, td1)
	fileStore.Replace(id2, td2)

	// query returns 2 sensors. not sure about the sort order
	res, err := fileStore.Query(queryString, 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))
	item1 := res[0].(map[string]interface{})
	item1ID := item1["id"]
	assert.Equal(t, id1, item1ID)
	// assert.Equal(t, res[0], td1)

	fileStore.Close()

}

func TestQueryValueProp(t *testing.T) {
	queryString := "$[?(@.properties..title=='The switch')]"
	id1 := "thing1"
	td1 := td.CreateTD(id1, vocab.DeviceTypeSensor)
	titleProp := td.CreateProperty("Title", "Device title", vocab.PropertyTypeAttr)
	td.AddTDProperty(td1, "title", titleProp)
	valueProp := td.CreateProperty("value", "Sensor value", vocab.PropertyTypeSensor)
	td.AddTDProperty(td1, string(vocab.PropertyTypeSensor), valueProp)

	id2 := "thing2"
	td2 := make(map[string]interface{})
	td2["id"] = "thing2"
	td2["type"] = "type2"
	td2[vocab.WoTAtType] = "sensor"
	td2["actions"] = make(map[string]interface{})
	td2["properties"] = make(map[string]interface{})
	td.AddTDProperty(td2, "title", "The switch")

	fileStore := makeFileStore()
	fileStore.Open()

	// dirstore.DirStoreCrud(t, fileStore)
	fileStore.Replace(id1, td1)
	fileStore.Replace(id2, td2)

	res, err := fileStore.Query(queryString, 0, 2)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	resJson, _ := json.MarshalIndent(res, " ", " ")
	fmt.Println(string(resJson))
	// logrus.Infof("query result: %s", resJson)

	fileStore.Close()

}
