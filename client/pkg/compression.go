package pkg

import (
	"bytes"
	"compress/gzip"
	"fileoptimizer/common" // common 패키지를 사용
	"fmt"
	"io/ioutil"
)

func CompressData(systemInfo interface{}, fileInfo []*common.FileInfo) ([]byte, error) { // common.FileInfo 사용
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)

	jsonData := []byte(fmt.Sprintf("SystemInfo: %v, FileInfo: %v", systemInfo, fileInfo))
	_, err := writer.Write(jsonData)
	if err != nil {
		return nil, err
	}

	writer.Close()
	return buffer.Bytes(), nil
}


// DecompressData는 gzip 압축된 데이터를 해제하는 함수입니다.
func DecompressData(compressedData []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 압축 해제된 데이터를 읽어서 반환
	decompressedData, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressedData, nil
}