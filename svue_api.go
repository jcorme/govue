package govue

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
)

type SVUEResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Result  string   `xml:"Body>ProcessWebServiceRequestResponse>ProcessWebServiceRequestResult"`
}

const (
	sVueEndpoint = "https://student-portland.cascadetech.org/portland/Service/PXPCommunication.asmx"
	soapAction   = "http://edupoint.com/webservices/ProcessWebServiceRequest"
)

func callApi(body io.Reader) (*SVUEResponse, error) {
	req, err := newSVueRequest(body)

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return decodeSVUEResponse(buf)
}

func newSVueRequest(body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", sVueEndpoint, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	return req, nil
}
