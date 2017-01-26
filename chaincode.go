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
	err = stub.PutState("rmids", blankBytes)
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
	} else if function == "createRM" {
		return t.createRM(stub, args)
	} else if function == "createBatch" {
		return t.createBatch(stub, args)
	} else if function == "createFoodPack" {
		return t.createFoodPack(stub, args)
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
	} else if function == "getAllRM" {
		return t.getAllRM(stub, args)
	}

	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// Peer one functions

type Cattle struct {
	Species     string  `json:"species"`
	CattleType  string  `json:"cattletype"`
	CattleId    string  `json:"cattleid"`
	CattleTag   string  `json:"cattletag"`
	Birthdate   string  `json:"birthdate"`
	Weight      float64 `json:"weight"`
	FarmerId    string  `json:"farmerid"`
	Status      string  `json:"status"`
	Certificate string  `json:"certificate"`
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
		Species:     args[0],
		CattleType:  args[1],
		CattleId:    args[2],
		CattleTag:   args[3],
		Birthdate:   args[4],
		Weight:      weight,
		FarmerId:    args[6],
		Status:      args[7],
		Certificate: args[11],
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
	prevHash := "\"prevHash\":\"" + args[10] + "\""

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

type Batch struct {
	Batchid   string   `json:"batchid"`
	Taglist   []string `json:"taglist"`
	Batchhdr  string   `json:"batchhdr"`
	Batchdate string   `json:"batchdate"`
	Source    string   `json:"source"`
	SourceHdr string   `json:"sourcehdr"`
}

type BatchList struct {
	Batch []string
}

// Create Batch
func (t *SimpleChaincode) createBatch(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// args 0 = "Farmername", 1 = "Batchid" , 2 = "[\"T01\",\"T02\"]" list of tags , 3 = batch date, 3 = sourcehdr

	batchkey := "batchids-" + args[0]

	batchidbytes, _ := stub.GetState(batchkey)

	// Create/update soruce's Batch List
	var batchlist BatchList

	err := json.Unmarshal(batchidbytes, &batchlist)

	batchlist.Batch = append(batchlist.Batch, args[1])

	batchidbytes, err = json.Marshal(batchlist)

	logger.Debug(batchlist)

	err = stub.PutState(batchkey, batchidbytes)

	var tags string
	tags = args[2]

	list := []string{tags}

	// Create Batch

	batch := Batch{
		Batchid:   args[1],
		Taglist:   list,
		Batchhdr:  batchkey,
		Batchdate: args[3],
		Source:    args[0],
		SourceHdr: args[4],
	}

	logger.Info(list)

	bytes, _ := json.Marshal(&batch)

	err = stub.PutState(batch.Batchid, bytes)

	for _, tag := range list {
		// Notify Each tag when the block is created
		var blank []string
		blank[1] = batch.SourceHdr + tag
		blank[2] = args[5]
		logger.Info(tag)
		t.updateHdr(stub, blank)
	}

	if err != nil {
		return nil, errors.New("Corrupt Transaction record")
	}

	return nil, nil
}

// Create Cattle Transfer
type TransferDetail struct {
	Transfer []string `json:"transfer"`
}

func (t *SimpleChaincode) createCattleTransfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// Creat or update Transaction in From side
	var transferFromdetails TransferDetail

	transferFrombytes, err := stub.GetState(args[3])

	err = json.Unmarshal(transferFrombytes, &transferFromdetails)

	transferFromdetails.Transfer = append(transferFromdetails.Transfer, args[0])
	transferFrombytes, err = json.Marshal(transferFromdetails)
	err = stub.PutState(args[3], transferFrombytes)

	// Creat or update Transaction in To side
	var transferTodetails TransferDetail
	transferTobytes, err := stub.GetState(args[3])

	err = json.Unmarshal(transferTobytes, &transferTodetails)

	transferTodetails.Transfer = append(transferTodetails.Transfer, args[0])
	transferTobytes, err = json.Marshal(transferTodetails)
	err = stub.PutState(args[4], transferTobytes)

	if err != nil {
		return nil, errors.New("Corrupt Transaction record")
	}

	return t.updateHdr(stub, args)

}

func (t *SimpleChaincode) updateHdr(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// Create and Update Cattle Header, Rawmeat and food pkg
	hdr := args[1]

	bytes, err := stub.GetState(hdr)

	if err != nil {
		return nil, errors.New("Corrupt Transaction record")
	}

	var headers CattleHeader

	err = json.Unmarshal(bytes, &headers)
	headers.Blockheader = append(headers.Blockheader, args[2])

	bytes, err = json.Marshal(headers)
	err = stub.PutState(hdr, bytes)

	return nil, nil
}

type Rawmeat struct {
	RawmeatId   string  `json:"rawmeatid"`
	Weight      float64 `json:"weight"`
	CreatedDate string  `json:"createddate"`
	SourceTag   string  `json:"sourcetag"`
	ExpireDate  string  `json:"expiredate"`
	Temperature string  `json:"temperature"`
	Company     string  `json:"company"`
	Certificate string  `json:"certificate"`
}

type Slaughter struct {
	Rawmeat []string `json:"rawmeats"`
}

// Get all cattle
func (t *SimpleChaincode) getAllRM(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp string
	var err error

	valAsbytes, err := stub.GetState("rmids")
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for rmids \"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// Peer Two function
func (t *SimpleChaincode) createRM(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Initializing Raw meat Creation")

	weight, err := strconv.ParseFloat(args[1], 64)

	if args[6] != SLAUGHTERHOUSE { // Only the farmer can create a cattle
		return nil, errors.New(fmt.Sprintf("Permission Denied. Create rawmeat . %v === %v", args[6], SLAUGHTERHOUSE))
	}

	rawmeat := Rawmeat{
		RawmeatId:   args[0],
		Weight:      weight,
		CreatedDate: args[2],
		SourceTag:   args[3],
		ExpireDate:  args[4],
		Temperature: args[5],
		Company:     args[6],
		Certificate: args[7],
	}

	bytes, err := json.Marshal(&rawmeat)

	if err != nil {
		return nil, err
	}

	err = stub.PutState(rawmeat.RawmeatId, bytes)

	if err != nil {
		return nil, err
	}

	bytes, err = stub.GetState("rmids")

	if err != nil {
		return nil, errors.New("Unable to get rmids")
	}

	// Create Cattle List
	var rawmeats Slaughter

	err = json.Unmarshal(bytes, &rawmeats)

	if err != nil {
		return nil, errors.New("Corrupt Farmer record")
	}

	rawmeats.Rawmeat = append(rawmeats.Rawmeat, rawmeat.RawmeatId)

	bytes, err = json.Marshal(rawmeats)

	err = stub.PutState("rmids", bytes)

	if err != nil {
		return nil, errors.New("Unable to put the state")
	}
	// Create Empty Blockheader list
	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	var cattletaghdr string

	cattletaghdr = "rawmeathdr-" + args[0]
	// Create Block Header json
	headerBlock := "\"block\":\"" + args[8] + "\", " // Variables to define the JSON
	headerType := "\"type\":\"RAWMEATCREATION\", "
	headerValue := "\"value\":\"" + args[0] + "\", "
	prevHash := "\"prevHash\":\"" + args[9] + "\""

	headerjson := "{" + headerBlock + headerType + headerValue + prevHash + "}"

	// save Blockheader
	var cattleheaders CattleHeader

	err = json.Unmarshal(blankBytes, &cattleheaders)
	cattleheaders.Blockheader = append(cattleheaders.Blockheader, headerjson)

	bytes, err = json.Marshal(cattleheaders)
	err = stub.PutState(cattletaghdr, bytes)

	var arguments []string
	arguments[1] = "cattlehdr-" + rawmeat.SourceTag
	arguments[2] = headerjson
	t.updateHdr(stub, arguments)

	return nil, nil
}

type FoodPack struct {
	Foodpackid          string  `json:"foodpackid"`
	Weight              float64 `json:"weight"`
	CreatedDate         string  `json:"createddate"`
	SourceTag           string  `json:"sourcetag"`
	ExpireDate          string  `json:"expiredate"`
	Temperature         string  `json:"temperature"`
	Company             string  `json:"company"`
	PerservationProcess string  `json:"perservationprocess"`
	Certificate         string  `json:"certificate"`
	PackageType         string  `json:"packagetype"`
	Productstate        string  `json:"productstate"`
	Primalcut           string  `json:"partname"`
}

type Foodmfg struct {
	FoodPack []string `json:"foodpack"`
}

func (t *SimpleChaincode) createFoodPack(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Initializing Food pack Creation")

	weight, err := strconv.ParseFloat(args[1], 64)

	if args[6] != MANUFACTURER { // Only the farmer can create a cattle
		return nil, errors.New(fmt.Sprintf("Permission Denied. Create food pack . %v === %v", args[6], MANUFACTURER))
	}

	foodpack := FoodPack{
		Foodpackid:          args[0],
		Weight:              weight,
		CreatedDate:         args[2],
		SourceTag:           args[3],
		ExpireDate:          args[4],
		Temperature:         args[5],
		Company:             args[6],
		PerservationProcess: args[7],
		Certificate:         args[8],
		PackageType:         args[9],
		Productstate:        args[10],
		Primalcut:           args[11],
	}

	bytes, err := json.Marshal(&foodpack)

	if err != nil {
		return nil, err
	}

	err = stub.PutState(foodpack.Foodpackid, bytes)

	if err != nil {
		return nil, err
	}

	bytes, err = stub.GetState("foodpackids")

	if err != nil {
		return nil, errors.New("Unable to get foodpackid")
	}

	// Create food List
	var foodpacks Foodmfg

	err = json.Unmarshal(bytes, &foodpacks)

	if err != nil {
		return nil, errors.New("Corrupt food pkg record")
	}

	foodpacks.FoodPack = append(foodpacks.FoodPack, foodpack.Foodpackid)

	bytes, err = json.Marshal(foodpacks)

	err = stub.PutState("foodpackids", bytes)

	if err != nil {
		return nil, errors.New("Unable to put the state")
	}
	// Create Empty Blockheader list
	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	var taghdr string

	taghdr = "foodpackhdr-" + args[0]
	// Create Block Header json
	headerBlock := "\"block\":\"" + args[12] + "\", " // Variables to define the JSON
	headerType := "\"type\":\"PACKAGING\", "
	headerValue := "\"value\":\"" + args[13] + "\", "
	prevHash := "\"prevHash\":\"" + args[14] + "\""

	headerjson := "{" + headerBlock + headerType + headerValue + prevHash + "}"

	// save Blockheader
	var cattleheaders CattleHeader

	err = json.Unmarshal(blankBytes, &cattleheaders)
	cattleheaders.Blockheader = append(cattleheaders.Blockheader, headerjson)

	bytes, err = json.Marshal(cattleheaders)
	err = stub.PutState(taghdr, bytes)

	var arguments []string
	arguments[0] = ""
	arguments[1] = "rawmeathdr-" + foodpack.SourceTag
	arguments[2] = headerjson

	t.updateHdr(stub, arguments) // update raw meat header

	var arguments1 []string
	arguments1[0] = ""
	arguments1[1] = args[15]
	arguments1[2] = headerjson

	t.updateHdr(stub, arguments1) // update cattle header

	return nil, nil
}
