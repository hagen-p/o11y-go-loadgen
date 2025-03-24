package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func DumpAndPauseRequest(req *http.Request, rawBody []byte) (*http.Request, error) {
	fmt.Println("==== Outgoing OTLP Request Headers ====")
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	fmt.Println("\n==== Outgoing OTLP Request Body ====")
	if isJSONContent(req.Header.Get("Content-Type")) {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, rawBody, "", "  "); err == nil {
			fmt.Println(pretty.String())
		} else {
			fmt.Println("⚠️ JSON indentation failed — showing raw:")
			fmt.Println(string(rawBody))
		}
	} else {
		fmt.Println("Raw Body (non-JSON):")
		fmt.Printf("%x\n", rawBody)
	}

	fmt.Print("\n[Paused] Press ENTER to continue...")
	_, _ = fmt.Fscanf(os.Stdin, "\n")

	// Rewrap the body for sending
	newReq := req.Clone(req.Context())
	newReq.Body = io.NopCloser(bytes.NewReader(rawBody))
	return newReq, nil
}

func isJSONContent(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}
