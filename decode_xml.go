package govue

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

type SVUERespError struct {
	XMLName xml.Name `xml:"RT_ERROR"`
	Message string   `xml:"ERROR_MESSAGE,attr"`
}

// Codes are:
//   0: StudentVue Server Error
//   1: Unexpected Error
//   2: Invalid Credentials
//   3: Response Decoding Error
type SVUEError struct {
	OrigError error

	Code int
}

const (
	SVueServerError = iota
	UnexpectedError
	InvalidCredentialsError
	DecodingError
)

const (
	sVueServerErrorMsg         = "An error has occurred with the StudentVue server."
	unexpectedErrorMsg         = "An unexpected error has occurred."
	invalidCredentialsErrorMsg = "The username and/or password is invalid."
	decodingErrorMsg           = "An internal error has occurred."
)

func (s SVUEError) Error() string {
	switch s.Code {
	case SVueServerError:
		return sVueServerErrorMsg
	case UnexpectedError:
		return unexpectedErrorMsg
	case InvalidCredentialsError:
		return invalidCredentialsErrorMsg
	case DecodingError:
		return decodingErrorMsg
	default:
		return unexpectedErrorMsg
	}
}

func decodeSVUEResponse(body *bytes.Buffer) (*SVUEResponse, error) {
	sVueResp := new(SVUEResponse)

	if err := xml.Unmarshal(body.Bytes(), sVueResp); err != nil {
		return nil, SVUEError{err, DecodingError}
	}

	return sVueResp, nil
}

func decodeStudentSignIn(sVueResp *SVUEResponse) (*Student, error) {
	resp := new(SVUESignInResponse)
	d, err := respIsOk(sVueResp, "ChildList")

	if err != nil {
		return nil, err
	}

	if err = d.Decode(resp); err != nil {
		return nil, SVUEError{err, DecodingError}
	}

	return resp.Students[0], nil
}

func decodeStudentGrades(sVueResp *SVUEResponse) (*Gradebook, error) {
	gb := new(Gradebook)
	d, err := respIsOk(sVueResp, "Gradebook")

	if err != nil {
		return nil, err
	}

	if err = d.Decode(gb); err != nil {
		return nil, SVUEError{err, DecodingError}
	}

	for _, c := range gb.Courses {
		c.CurrentMark = c.Marks[gb.CurrentGradingPeriodIndex()]
	}

	return gb, nil
}

func respIsOk(sVueResp *SVUEResponse, expectedElement string) (*xml.Decoder, error) {
	d := xml.NewDecoder(strings.NewReader(sVueResp.Result))

TokenLoop:
	for {
		t, err := d.Token()

		if err == io.EOF {
			return nil, SVUEError{err, DecodingError}
		}

		if _t, ok := t.(xml.StartElement); ok {
			switch _t.Name.Local {
			case expectedElement:
				break TokenLoop
			case "RT_ERROR":
				return nil, decodeRespError(sVueResp)
			default:
				continue TokenLoop
			}
		} else {
			continue
		}
	}

	return xml.NewDecoder(strings.NewReader(sVueResp.Result)), nil
}

func decodeRespError(sVueResp *SVUEResponse) error {
	sErr := new(SVUERespError)

	err := xml.Unmarshal([]byte(sVueResp.Result), sErr)

	if err != nil {
		return SVUEError{err, DecodingError}
	}

	switch {
	case strings.Contains(sErr.Message, "The user name or password is incorrect."):
		return SVUEError{nil, InvalidCredentialsError}
	case strings.Contains(sErr.Message, "Invalid user id or password"):
		return SVUEError{nil, InvalidCredentialsError}
	default:
		return SVUEError{nil, UnexpectedError}
	}
}
