package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	scalecodec "github.com/itering/scale.go"
	"github.com/itering/scale.go/types"
	"github.com/itering/scale.go/utiles"
)

//load metadata from the downlaoded files
func getMetaForSpecVersion(specVersion int) (*types.MetadataStruct, *string, error) {
	metaPath := os.Getenv("METADATA_FILES_PARENT")

	fullPath := path.Join(metaPath, strconv.Itoa(specVersion))
	rawMeta, err := ioutil.ReadFile(fullPath)
	if err != nil {
		fmt.Println("Error reading metadata file for spec version", specVersion, err)
		return nil, nil, err
	}
	rawString := string(rawMeta)

	m := scalecodec.MetadataDecoder{}
	m.Init(utiles.HexToBytes(rawString))
	err = m.Process()
	if err != nil {
		fmt.Println("Error processing metadata for spec version", specVersion, err)
		return nil, nil, err
	}

	return &m.Metadata, &rawString, err
}
