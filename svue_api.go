package govue

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SVUEResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Result  string   `xml:"Body>ProcessWebServiceRequestResponse>ProcessWebServiceRequestResult"`
}

type SVUESignInResponse struct {
	XMLName  xml.Name   `xml:"ChildList"`
	Students []*Student `xml:"Child"`
}

const (
	sVueEndpoint = "https://student-portland.cascadetech.org/portland/Service/PXPCommunication.asmx"
	soapAction   = "http://edupoint.com/webservices/ProcessWebServiceRequest"
)

const (
	signInRequestBody = `<?xml version="1.0" encoding="utf-8"?>
		<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
			<soap:Body>
				<ProcessWebServiceRequest xmlns="http://edupoint.com/webservices/">
					<userID>%s</userID>
					<password>%s</password>
					<skipLoginLog>0</skipLoginLog>
					<parent>0</parent>
					<webServiceHandleName>PXPWebServices</webServiceHandleName>
					<methodName>ChildList</methodName>
					<paramStr/>
				</ProcessWebServiceRequest>
			</soap:Body>
		</soap:Envelope>`
	getGradesRequestBody = `<?xml version="1.0" encoding="utf-8"?>
		<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
			<soap:Body>
				<ProcessWebServiceRequest xmlns="http://edupoint.com/webservices/">
					<userID>%s</userID>
					<password>%s</password>
					<skipLoginLog>1</skipLoginLog>
					<parent>0</parent>
					<webServiceHandleName>PXPWebServices</webServiceHandleName>
					<methodName>Gradebook</methodName>
					<paramStr>&lt;Parms&gt;&lt;ChildIntID&gt;0&lt;/ChildIntID&gt;&lt;/Parms&gt;</paramStr>
				</ProcessWebServiceRequest>
			</soap:Body>
		</soap:Envelope>`
)

func SignInStudent(username, password string) (*Student, error) {
	signInBody := fmt.Sprintf(signInRequestBody, username, password)
	sResp, err := callApi(strings.NewReader(signInBody))

	if err != nil {
		return nil, err
	}

	return decodeStudentSignIn(sResp)
}

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
