package govue

import (
	"bytes"
	"encoding/xml"
)

func decodeSVUEResponse(body *bytes.Buffer) (*SVUEResponse, error) {
	sVueResp := new(SVUEResponse)

	if err := xml.Unmarshal(body.Bytes(), sVueResp); err != nil {
		return nil, err
	}

	return sVueResp, nil
}
