package main

import (
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestBatchPromoCSVFile(t *testing.T) {
	promoCsvContent := `172FFC14-D229-4C93-B06B-F48B8C095511,9.68,"2022-06-04 06:01:21"
172FFC14-D229-4C93-B06B-F48B8C095512,9.68,"2022-06-04 06:01:22"
172FFC14-D229-4C93-B06B-F48B8C095513,9.68,"2022-06-04 06:01:23"
172FFC14-D229-4C93-B06B-F48B8C095514,9.68,"2022-06-04 06:01:24"
172FFC14-D229-4C93-B06B-F48B8C095515,9.68,"2022-06-04 06:01:25"
`

	myReader := strings.NewReader(promoCsvContent)
	noOfLines, err := lineCounter(myReader)
	assert.Equal(t, err, nil)
	assert.Equal(t, 5, noOfLines)

	myReader = strings.NewReader(promoCsvContent)
	promoList := batchPromoRecordsFromCSV(myReader, 2, 4)
	assert.Equal(t, len(promoList), 3)
	assert.Equal(t, promoList[0].ID, "172FFC14-D229-4C93-B06B-F48B8C095512")
	assert.Equal(t, promoList[1].ID, "172FFC14-D229-4C93-B06B-F48B8C095513")
	assert.Equal(t, promoList[2].ID, "172FFC14-D229-4C93-B06B-F48B8C095514")

	myReader = strings.NewReader(promoCsvContent)
	promoList = batchPromoRecordsFromCSV(myReader, 4, 4)
	assert.Equal(t, len(promoList), 1)
	assert.Equal(t, promoList[0].ID, "172FFC14-D229-4C93-B06B-F48B8C095514")

	myReader = strings.NewReader(promoCsvContent)
	promoList = batchPromoRecordsFromCSV(myReader, 1, 1)
	assert.Equal(t, len(promoList), 1)
	assert.Equal(t, promoList[0].ID, "172FFC14-D229-4C93-B06B-F48B8C095511")

	myReader = strings.NewReader(promoCsvContent)
	promoList = batchPromoRecordsFromCSV(myReader, 0, 1)
	assert.Equal(t, len(promoList), 1)
	assert.Equal(t, promoList[0].ID, "172FFC14-D229-4C93-B06B-F48B8C095511")

	myReader = strings.NewReader(promoCsvContent)
	promoList = batchPromoRecordsFromCSV(myReader, 1, 2)
	assert.Equal(t, len(promoList), 2)
	assert.Equal(t, promoList[0].ID, "172FFC14-D229-4C93-B06B-F48B8C095511")
	assert.Equal(t, promoList[1].ID, "172FFC14-D229-4C93-B06B-F48B8C095512")

	myReader = strings.NewReader(promoCsvContent)
	promoList = batchPromoRecordsFromCSV(myReader, 4, 7)
	assert.Equal(t, len(promoList), 2)
	assert.Equal(t, promoList[0].ID, "172FFC14-D229-4C93-B06B-F48B8C095514")
	assert.Equal(t, promoList[1].ID, "172FFC14-D229-4C93-B06B-F48B8C095515")
}
