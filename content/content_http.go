package content

import (
	"net/http"
	"fmt"
	"os"
	"time"
	"strconv"
	"io"
	"io/ioutil"
	"bytes"
	"encoding/json"
	"encoding/base64"
	//"encoding/hex"
	shell "github.com/ipfs/go-ipfs-api"
)

var UpstreamIPFSAddress string

func HttpObjectGet(rw http.ResponseWriter, req *http.Request) (string, bool) {

	var multihash = req.URL.Query().Get("arg")
	fmt.Printf("Fetching %s...\n", multihash)

	url := fmt.Sprintf("%s%s", UpstreamIPFSAddress, req.URL.String())

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
	    Timeout: timeout,
	}
	ipfsResponse, err := client.Get(url)

	if err != nil || ipfsResponse.StatusCode != 200 {
		var error_string = fmt.Sprintf("{\"error\":\"%s\"}", err)
		rw.WriteHeader(http.StatusNotFound)
		rw.Header().Set("Content-Length", strconv.Itoa(len(error_string)))
		fmt.Fprintf(rw, "%s", error_string)
		fmt.Printf("Fetching %s... Not found.\n", multihash)
		rw.(http.Flusher).Flush()
		req.Close = true

		return multihash, false
	}

	fmt.Printf("Found %s.\n", multihash)
	io.Copy(rw, ipfsResponse.Body)
	rw.(http.Flusher).Flush()

	return multihash, true
}

func HttpShutdown(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Length", "14")
	fmt.Fprintf(rw, "Shutting down\n")
	fmt.Printf("Shutting down\n")
	rw.(http.Flusher).Flush()
	os.Exit(0)
}

func HaveContent(multihash string) bool {

	url := fmt.Sprintf("%s%s", UpstreamIPFSAddress, "/api/v0/refs?arg=" + multihash)

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
	    Timeout: timeout,
	}
	ipfsResponse, err := client.Get(url)

	if err != nil {
		//fmt.Printf("HaveContent Error %s\n", err)
		return false
	}

	body, err := ioutil.ReadAll(ipfsResponse.Body)
	bodyStr := bytes.NewBuffer(body)

	defer ipfsResponse.Body.Close()

	if bodyStr.String() == "Error: merkledag: not found" {
		return false
	} else {
		return true
	}
}

func Add(hash string, content_b64 string) bool {

	decoded, _ := base64.StdEncoding.DecodeString(content_b64)

	//fmt.Printf("Content Length %v\n", len(decoded))

	var tmpObject shell.IpfsObject

	tmpObject = shell.IpfsObject{}
	err := json.Unmarshal(decoded, &tmpObject)

	bIsDAG := true
	if tmpObject.Data == "" && tmpObject.Links == nil {
		bIsDAG = false
	}

	var res string

	myshell := shell.NewShell(UpstreamIPFSAddress)
	if bIsDAG {
		res, err = myshell.ObjectPut(&tmpObject)

		if err != nil {
			fmt.Printf("Error with IPFS Object/Put %s\n", err)
			return false
		}
	} else {

		var decodedBuf = bytes.NewBuffer(decoded)
		//if not a DAG
		res, err = myshell.Add(decodedBuf)

		if err != nil {
			fmt.Printf("Error with IPFS Add %s\n", err)
			return false
		}
	}

	if res != hash {
		fmt.Printf("IPFS Add expected %s, response = %s\n\n", hash, res)
		return false
	}

	fmt.Printf("Content.Add: Successfully Added %s\n", hash)

	return true
}

//returns a base64 encoded string
func Get(hash string) string {

	var res1 io.Reader
	var res2 *shell.IpfsObject
	isFile := true
	var err error

	myshell := shell.NewShell(UpstreamIPFSAddress)
	res1, err = myshell.Cat(hash)

	if err != nil {
		//fmt.Printf("Error in content.Get: %s", err)
		//could not cat the file because it was a IPFS Object

		//now try object get
		res2, err = myshell.ObjectGet(hash)

		if err != nil {
			fmt.Printf("content.Get, Could not Cat or ObjectGet hash: %s\n", err)
			return ""
		}

		isFile = false
	}

	var bodyBytes []byte

	if isFile {
		bodyBytes, _ = ioutil.ReadAll(res1)
	} else {
		bodyBytes, _ = json.Marshal(res2)
	}

	b64 := base64.StdEncoding.EncodeToString(bodyBytes)

	return b64
}

