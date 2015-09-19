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

func decodeStudentSignIn(sVueResp *SVUEResponse) (*Student, error) {
	resp := new(SVUESignInResponse)

	if err := xml.Unmarshal([]byte(sVueResp.Result), resp); err != nil {
		return nil, err
	}

	return resp.Students[0], nil
}
