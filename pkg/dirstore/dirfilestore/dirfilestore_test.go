package dirfilestore_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ohler55/ojg/jp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/wostdir/pkg/dirstore"
	"github.com/wostzone/wostdir/pkg/dirstore/dirfilestore"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/vocab"
)

func makeFileStore() *dirfilestore.DirFileStore {
	filename := "/tmp/test-dirfilestore.json"

	store := dirfilestore.NewDirFileStore(filename)
	return store
}

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
	td2 := td.CreateTD(id2, vocab.DeviceTypeOnOffSwitch)
	td.AddTDProperty(td2, "title", "The switch")
	fileStore := makeFileStore()
	// dirstore.DirStoreCrud(t, fileStore)
	tdd := map[string]interface{}(td1)
	fileStore.Create(id1, tdd)
	tdd = map[string]interface{}(td2)
	fileStore.Create(id2, tdd)
	// fileStore.Close()

	t1 := time.Now()
	var i int
	for i = 0; i < 1; i++ {

		// regular filter
		res, err := fileStore.Query(`$[?(@.id=="thing1")]`)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// regular nested filter comparison
		res, err = fileStore.Query(`$[?(@.properties.title=="The sensor")]`)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with nested notation
		res, err = fileStore.Query(`$.*[?(@.title=="The sensor")]`)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with bracket notation
		res, err = fileStore.Query(`$[?(@["id"]=="thing1")]`)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with bracket notation and current object literal (for search @type)
		res, err = fileStore.Query(`$[?(@['@type']=="sensor")]`)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	}
	d1 := time.Since(t1)
	logrus.Infof("TestQuery, %d runs: %d msec", i, d1.Milliseconds())
	// logrus.Infof("Query Results:\n%v", res)
}

func TestQueryBracketNotation(t *testing.T) {
	store := make(map[string]interface{})

	jsondoc := `{
		"item1": {
			"id": "item1",
			"type": "type1",
			"@type": "attype1"
		},
		"item2": {
			"id": "item2",
			"type": "type2",
			"@type": "attype2"
		}
	}`
	query1 := `$[?(@['type']=="type1")]`
	query2 := `$[?(@['@type']=="type1")]`

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
