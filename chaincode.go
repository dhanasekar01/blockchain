package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const AUTHORITY = "regulator"
const MANUFACTURER = "manufacturer"
const FARMER = "farmer"
const RETAILER = "walmart"
const SLAUGHTERHOUSE = "slaughterhouse"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

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
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "createCattle" {
		return t.createCattle(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	} else if function == "getCattle" {
		return t.getCattle(stub, args)
	} else if function == "getAllCattle" {
		return t.getAllCattle(stub, args)
	}

	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// write - invoke function to write key/value pair
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
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

	err = json.Unmarshal(bytes, &cattletag)

	if cattletag != "" {
		return nil, errors.New(fmt.Sprintf("Cattle Already Present"))
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

	var cattles Farmer

	err = json.Unmarshal(bytes, &cattles)

	if err != nil {
		return nil, errors.New("Corrupt Farmer record")
	}

	cattles.Cattle = append(cattles.Cattle, cattle.CattleTag)

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

// read cattle
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

// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}
