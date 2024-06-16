package qgis

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type Qgis struct {
	token string
}

func NewQgis(token string) *Qgis {
	return &Qgis{
		token: token,
	}
}

func (q *Qgis) GetSatelliteData(x1 string, y1 string, x2 string, y2 string) (string, error) {
	// The URL for the API endpoint
	url := "https://services.sentinel-hub.com/api/v1/process"
	// Create a new buffer to store the form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the request JSON
	requestJson := fmt.Sprintf(`{
        "input": {
            "bounds": {
                "properties": {
                    "crs": "http://www.opengis.net/def/crs/EPSG/0/32633"
                },
                "bbox": [
                    %s,
                    %s,
                   	%s,
                    %s
                ]
            },
            "data": [
                {
                    "type": "sentinel-2-l2a",
                    "dataFilter": {
                        "timeRange": {
                            "from": "2022-10-01T00:00:00Z",
                            "to": "2022-10-31T00:00:00Z"
                        }
                    }
                }
            ]
        },
        "output": {
            "width": 512,
            "height": 512
        }
    }`, x1, y1, x2, y2)
	requestPart, err := writer.CreateFormField("request")
	if err != nil {
		return "", errors.New("Error creating form field: " + err.Error())
	}
	io.Copy(requestPart, strings.NewReader(requestJson))

	// Add the evalscript
	evalscript := `//VERSION=3
	function setup() {
	  return {
		input: ["B02", "B03", "B04"],
		output: { bands: 3 }
	  }
	}
	function evaluatePixel(sample) {
	  return [2.5 * sample.B04, 2.5 * sample.B03, 2.5 * sample.B02]
	}`

	evalscriptPart, err := writer.CreateFormField("evalscript")
	if err != nil {
		return "", errors.New("Error creating form field: " + err.Error())
	}
	io.Copy(evalscriptPart, strings.NewReader(evalscript))

	// Close the writer to finalize the form data
	writer.Close()

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return "", errors.New("Error creating request: " + err.Error())
	}

	// Add the authorization header
	req.Header.Add("Authorization", q.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("Error making request: " + err.Error())
	}
	defer resp.Body.Close()

	// Print the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("Error reading response: " + err.Error())
	}
	return string(body), nil
}
