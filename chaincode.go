package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("CLDChaincode")

const AUTHORITY = "regulator"
const MANUFACTURER = "manufacturer"
const FARMER = "farmer"
const RETAILER = "walmart"
const SLAUGHTERHOUSE = "slaughterhouse"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

	username, err := stub.ReadCertAttribute("username")
	if err != nil {
		return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error())
	}
	return string(username), nil
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Initializing cattle id collection")

	var blank []string
	blankBytes, _ := json.Marshal(&blank)

	err := stub.PutState("cattleids", blankBytes)
	if err != nil {
		fmt.Println("Failed to initialize cattle Id collection")
	}

	fmt.Println("Initialization complete")
	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createCattle" {
		return t.createCattle(stub, args)
	} else if function == "createCattleTransfer" {
		return t.createCattleTransfer(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "getCattle" {
		return t.getCattle(stub, args)
	} else if function == "getAllCattle" {
		return t.getAllCattle(stub, args)
	} else if function == "getCattleTrans" {
		return t.getCattleTrans(stub, args)
	}

	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// Peer one functions

type Cattle struct {
	Species    string  `json:"species"`
	CattleType string  `json:"cattletype"`
	CattleId   string  `json:"cattleid"`
	CattleTag  string  `json:"cattletag"`
	Birthdate  string  `json:"birthdate"`
	Weight     float64 `json:"weight"`
	FarmerId   string  `json:"farmerid"`
	Status     string  `json:"status"`
}

type Farmer struct {
	Cattle []string `json:"cattle"`
}

type CattleHeader struct {
	Blockheader []string `json:"blockheader"`
}

func (t *SimpleChaincode) createCattle(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var cattletag string

	fmt.Println("Initializing Cattle Creation")

	weight, err := strconv.ParseFloat(args[5], 64)

	if args[6] != FARMER { // Only the farmer can create a cattle
		return nil, errors.New(fmt.Sprintf("Permission Denied. Create Cattle. %v === %v", args[6], FARMER))
	}

	bytes, err := stub.GetState(args[3])

	if bytes != nil {
		err = json.Unmarshal(bytes, &cattletag)

		if cattletag != "" {
			return nil, errors.New(fmt.Sprintf("Cattle Already Present"))
		}
	}

	cattle := Cattle{
		Species:    args[0],
		CattleType: args[1],
		CattleId:   args[2],
		CattleTag:  args[3],
		Birthdate:  args[4],
		Weight:     weight,
		FarmerId:   args[6],
		Status:     args[7],
	}

	bytes, err = json.Marshal(&cattle)

	if err != nil {
		return nil, err
	}

	err = stub.PutState(cattle.CattleTag, bytes)

	if err != nil {
		return nil, err
	}

	bytes, err = stub.GetState("cattleids")

	if err != nil {
		return nil, errors.New("Unable to get cattleids")
	}

	// Create Cattle List
	var cattles Farmer

	err = json.Unmarshal(bytes, &cattles)

	if err != nil {
		return nil, errors.New("Corrupt Farmer record")
	}

	cattles.Cattle = append(cattles.Cattle, cattle.CattleTag)

	bytes, err = json.Marshal(cattles)

	err = stub.PutState("cattleids", bytes)

	if err != nil {
		return nil, errors.New("Unable to put the state")
	}
	// Create Empty Blockheader list
	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	var cattletaghdr string

	cattletaghdr = "cattlehdr-" + args[3]
	// Create Block Header json
	headerBlock := "\"block\":\"" + args[8] + "\", " // Variables to define the JSON
	headerType := "\"type\":\"Create\", "
	headerValue := "\"value\":\"" + args[9] + "\", "
	prevHash := "\"prevHash\":\"" + args[10] + "\", "

	headerjson := "{" + headerBlock + headerType + headerValue + prevHash + "}"

	// save Blockheader
	var cattleheaders CattleHeader

	err = json.Unmarshal(blankBytes, &cattleheaders)
	cattleheaders.Blockheader = append(cattleheaders.Blockheader, headerjson)

	bytes, err = json.Marshal(cattleheaders)
	err = stub.PutState(cattletaghdr, bytes)

	return nil, nil
}

// read cattle
func (t *SimpleChaincode) getCattle(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error
	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// Get all cattle
func (t *SimpleChaincode) getAllCattle(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp string
	var err error

	valAsbytes, err := stub.GetState("cattleids")
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for cattleids \"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// Get all Cattle Transaction
func (t *SimpleChaincode) getCattleTrans(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp string
	var err error

	valAsbytes, err := stub.GetState("cattlehdr-" + args[0])
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for Cattle Transactions \"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// Create Cattle Transfer

type Transfer struct {
	Id     string `json:"id"`
	Value  string `json:"value"`
	Header string `json:"header"`
	From   string `json:"from"`
	To     string `json:"to"`
	Date   string `json:"date"`
}

type TransferDetail struct {
	Transfer []string `json:"transfer"`
}

func (t *SimpleChaincode) createCattleTransfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	transfer := Transfer{
		Id:     args[0], // TAG01
		Value:  args[1], // cattlehdr
		Header: args[2], // Transfered to Typon Fresh Meat
		From:   args[3], // American Farmer
		To:     args[4], // Typon Fresh Meat
		Date:   args[5], // Date
	}

	Id := "\"id\":\"" + transfer.Id + "\", " // Variables to define the JSON
	Value := "\"value\":\"" + transfer.Value + "\", "
	Header := "\"header\":\"" + transfer.Header + "\", "
	From := "\"from\":\"" + transfer.From + "\", "
	To := "\"to\":\"" + transfer.To + "\", "
	Date := "\"date\":\"" + transfer.Date + "\","

	transferjson := "{" + Id + Value + Header + From + To + Date + "}"

	// Creat or update Transaction in From side
	var transferFromdetails TransferDetail

	// Create Empty Blockheader list
	var blank []string
	transferFrombytes, _ := json.Marshal(&blank)

	err = json.Unmarshal(transferFrombytes, &transferFromdetails)

	if err != nil {
		return nil, errors.New("Corrupt Transaction record")
	}

	transferFromdetails.Transfer = append(transferFromdetails.Transfer, transferjson)
	transferFrombytes, err = json.Marshal(transferFromdetails)
	err = stub.PutState(transfer.From, transferFrombytes)

	if err != nil {
		return nil, errors.New("Corrupt Transaction record")
	}

	// Creat or update Transaction in To side
	var transferTodetails TransferDetail
	transferTobytes, _ := json.Marshal(&blank)

	err = json.Unmarshal(transferTobytes, &transferTodetails)

	if err != nil {
		return nil, errors.New("Corrupt Transaction record")
	}

	transferTodetails.Transfer = append(transferTodetails.Transfer, transferjson)
	transferTobytes, err = json.Marshal(transferTodetails)
	err = stub.PutState(transfer.To, transferTobytes)

	if err != nil {
		return nil, errors.New("Corrupt Transaction record")
	}

	var arg []string
	arg[0] = transfer.Value
	arg[1] = transfer.Id
	arg[2] = args[6]
	arg[3] = args[7]
	arg[4] = args[8]
	arg[5] = args[9]

	return t.updateHdr(stub, arg)

}

func (t *SimpleChaincode) updateHdr(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// Create and Update Cattle Header
	hdr := args[0] + args[1]

	headerBlock := "\"block\":\"" + args[2] + "\", " // Variables to define the JSON
	headerType := "\"type\":\"" + args[3] + "\", "
	headerValue := "\"value\":\"" + args[4] + "\", "
	prevHash := "\"prevHash\":\"" + args[5] + "\", "

	headerjson := "{" + headerBlock + headerType + headerValue + prevHash + "}"

	bytes, err := stub.GetState(hdr)

	if err != nil {
		return nil, errors.New("Corrupt Transaction record")
	}

	var headers CattleHeader

	err = json.Unmarshal(bytes, &headers)
	headers.Blockheader = append(headers.Blockheader, headerjson)

	bytes, err = json.Marshal(headers)
	err = stub.PutState(hdr, bytes)

	return nil, nil
}
