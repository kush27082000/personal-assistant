package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Define a struct to parse Wit.ai responses
type WitAIResponse struct {
	Text string `json:"text"`
}

type WitAIReq struct {
	Message string `json:"message"`
}

func HandleWitAIRequest(c *gin.Context) {
	// Parse the user's message from the request

	message := WitAIReq{}
	if err := c.BindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}

	inputString := message.Message
	msg := strings.ReplaceAll(inputString, " ", "%20")

	// Define your Wit.ai API key and endpoint
	witAPIKey := WIT_AI_KEY
	witEndpoint := "https://api.wit.ai/message?q=" + msg

	// Create an HTTP client with your Wit.ai API key
	client := &http.Client{}
	req, err := http.NewRequest("GET", witEndpoint, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Authorization", "Bearer "+witAPIKey)

	// Send the message to Wit.ai for processing
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read and parse Wit.ai's response
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var witResponse WitAIResponse2
	err = json.Unmarshal(responseBody, &witResponse)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// You can process witResponse.Text and use it as the bot's response

	// For now, let's just return the response to the client
	// response := map[string]interface{}{
	// 	"message": "witResponse.Text2222",
	// 	// "message": witResponse,
	// }

	if witResponse.Entities.OrderType == nil || witResponse.Entities.SchemeName == nil || witResponse.Entities.AmountOfMoney == nil {
		response := map[string]interface{}{
			"message": "Hi, please enter your request. For example, you can ask me to place an order on your behalf.",
			// "message": witResponse,
		}

		c.JSON(http.StatusOK, response)
		return
	} else {

		amount := witResponse.Entities.AmountOfMoney[0].Value.(float64)
		myAmount := fmt.Sprintf("%.2f", amount)
		ordertype := witResponse.Entities.OrderType[0].Value.(string)
		schemename := witResponse.Entities.SchemeName[0].Value.(string)

		fmt.Printf("Order Type: %s\n", witResponse.Entities.OrderType[0].Value.(string))
		fmt.Printf("Scheme Name: %s\n", witResponse.Entities.SchemeName[0].Value.(string))
		fmt.Printf("Amount: %s\n", myAmount)
		content := "Okay, we know that you want to place a " + witResponse.Entities.OrderType[0].Value.(string) + " order of amount " + myAmount + " in a scheme of " + witResponse.Entities.SchemeName[0].Value.(string)
		// response := map[string]interface{}{
		// 	"message": content,
		// 	// "message": witResponse,
		// }
		// fmt.Printf("content %s\n", content)
		// c.JSON(http.StatusOK, response)

		// Define the API endpoint
		baseURL := "TBA PLEASE DEFINE URL"

		// Set the scheme_name parameter
		schemeName := witResponse.Entities.SchemeName[0].Value.(string) // Replace with your scheme_name

		// Encode the scheme_name
		queryParams := url.Values{}
		queryParams.Add("value", schemeName)
		queryParams.Add("limit", "5")

		// Create the full API URL with encoded query parameters
		apiURL := baseURL + "?" + queryParams.Encode()

		// Create an HTTP client
		client := &http.Client{}

		// Create a GET request
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		// Set request headers
		req.Header.Add("accept", "application/json")
		req.Header.Add("Authorization", MF_CORE_KEY)
		req.Header.Add("userType", "B2C")
		req.Header.Add("accountType", "D")

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		// Read the response body
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			response := map[string]interface{}{
				"message": "Sorry, your scheme name might be incorrect, please check your request again",
				// "message": witResponse,
			}
			c.JSON(http.StatusOK, response)
			return
		}

		// Print the response body
		fmt.Println(string(responseBody))

		// Parse the JSON response into the struct
		var apiResponse ApiResponse
		if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}

		// Access the scheme data
		schemes := apiResponse.Data.Schemes

		if len(schemes) == 0 {
			response := map[string]interface{}{
				"message": "Sorry, we could not find any scheme for your input, please try again with a different keyword",
				"schemes": schemes,
				// "message": witResponse,
			}
			c.JSON(http.StatusOK, response)
			return
		}

		// Print the scheme details
		for _, scheme := range schemes {
			fmt.Printf("Scheme Name: %s\n", scheme.SchemeName)
			fmt.Printf("Scheme Code: %s\n", scheme.SchemeCode)
			fmt.Printf("Category: %s\n", scheme.CategoryName)
			fmt.Printf("Subcategory: %s\n", scheme.SubcategoryName)
			fmt.Printf("Reinvestment Plan: %s\n", scheme.ReInvestmentPlan)
			fmt.Printf("Logo URL: %s\n", scheme.LogoUrl)
			fmt.Printf("3-Year Returns: %.2f\n", scheme.Returns3yr)
			fmt.Printf("ARQ Rating: %.2f\n", scheme.ARQRating)
			fmt.Println("------------------------------")
		}

		response := map[string]interface{}{
			"message":    content,
			"schemes":    schemes,
			"ordertype":  ordertype,
			"schemename": schemename,
			"amount":     amount,
			// "message": witResponse,
		}
		c.JSON(http.StatusOK, response)
		return
	}
}

var WIT_AI_KEY string
var MF_CORE_KEY string

func main() {

	WIT_AI_KEY = os.Getenv("WIT_AI_KEY")
	MF_CORE_KEY = os.Getenv("MF_CORE_KEY")
	println(MF_CORE_KEY)
	if WIT_AI_KEY == "" || MF_CORE_KEY == "" {
		panic("application won't function without WIT_AI_KEY and MF_CORE_KEY")
	}

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))

	r.POST("/witai", HandleWitAIRequest)
	r.POST("/scheme-details", HandleSchemeDetails)

	fmt.Println("Server is running on port 8080")
	r.Run(":8080")
}

type WitAIResponse2 struct {
	Entities Entities `json:"entities"`
	Intents  []Intent `json:"intents"`
	Text     string   `json:"text"`
	Traits   Traits   `json:"traits"`
}

type Entities struct {
	OrderType     []Entity `json:"order_type:order_type"`
	SchemeName    []Entity `json:"scheme_name:scheme_name"`
	AmountOfMoney []Entity `json:"wit$amount_of_money:amount_of_money"`
	// Add more entity types as needed
}

type Entity struct {
	Body       interface{} `json:"body"`
	Confidence float64     `json:"confidence"`
	End        int         `json:"end"`
	Entities   struct{}    `json:"entities"`
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Role       string      `json:"role"`
	Start      int         `json:"start"`
	Type       string      `json:"type"`
	Value      interface{} `json:"value"`
	Unit       string      `json:"unit,omitempty"`
}

type Intent struct {
	Confidence float64 `json:"confidence"`
	ID         string  `json:"id"`
	Name       string  `json:"name"`
}

type Traits struct {
}

type Scheme struct {
	SchemeCode       string  `json:"schemeCode"`
	ISIN             string  `json:"isin"`
	SchemeName       string  `json:"schemeName"`
	CategoryName     string  `json:"categoryName"`
	SubcategoryName  string  `json:"subcategoryName"`
	ReInvestmentPlan string  `json:"reInvestmentPlan"`
	LogoUrl          string  `json:"logoUrl"`
	Returns3yr       float64 `json:"returns3yr"`
	ARQRating        float64 `json:"arqRating"`
}

type Data struct {
	Schemes []Scheme `json:"schemes"`
}

type ApiResponse struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

func HandleSchemeDetails(c *gin.Context) {
	// Parse the scheme selection request from the frontend
	var request SchemeDetailsRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
		return
	}

	// Perform any necessary actions based on the scheme selection
	// For example, you can retrieve additional details from the API using the scheme code and ISIN

	// For demonstration purposes, we'll respond with the received scheme details
	// schemeDetails := SchemeDetailsResponse{
	// 	SchemeCode:       request.SchemeCode,
	// 	ISIN:             request.ISIN,
	// 	SchemeName:       "Sample Scheme Name",
	// 	CategoryName:     "Sample Category",
	// 	SubcategoryName:  "Sample Subcategory",
	// 	ReInvestmentPlan: "Sample Plan",
	// 	LogoUrl:          "https://example.com/sample-logo.png",
	// 	Returns3yr:       10.5,
	// 	ARQRating:        4,
	// }

	fmt.Println("we have scheme_code " + request.SchemeCode + " and isin " + request.ISIN + " schemename ordertype " + " " + request.SchemeName + "   " + request.OrderType)
	fmt.Println("amount   ")
	fmt.Print(request.Amount)

	// // c.JSON(http.StatusOK, schemeDetails)
	// response := map[string]interface{}{
	// 	"message": "yeah we have now ypur scheme code and isin and need amount of money some how and sip , we will figure out that",
	// 	// "schemes": schemes,
	// 	// "message": witResponse,
	// }
	// c.JSON(http.StatusOK, response)

	//-----------placing the order---------- //

	// Define the URL and request body data
	url := "TBA"

	// Create a SIP placement request with specific data
	currentDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	requestSIPPlacemet := SIPPlacementRequest{
		DpNumber:             "12345567788",
		EmandateId:           "",
		FirstOrderToday:      false,
		FolioNumber:          "",
		Frequency:            "MONTHLY",
		InstallmentAmount:    request.Amount,
		Isin:                 request.ISIN,
		NoOfInstallment:      99,
		SchemeCode:           request.SchemeCode,
		StartDate:            currentDate,
		TransactionRefNumber: "",
		Type:                 "SIP",
		VeryFirstSip:         false,
	}

	// Marshal the request into JSON
	requestData, err := json.Marshal(requestSIPPlacemet)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "SIP Placement failed due to Failed to marshal request data"})
		return
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	id := uuid.New()

	// Set request headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("X-Request-Id", id.String())
	req.Header.Set("Authorization", MF_CORE_KEY)
	req.Header.Set("userType", "B2C")
	req.Header.Set("accountType", "D")
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client
	client := &http.Client{}

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP Status Code: %d\n", resp.StatusCode)

		// Read the response body
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		fmt.Println("Response Body:", string(responseBody))

		// Parse the JSON response into the struct
		var errResponse ErrorResponse
		if err := json.Unmarshal(responseBody, &errResponse); err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "SIP placement error occured due to " + errResponse.Message})

		// // Parse the JSON response into the struct
		// var errResponse ErrorResponse
		// if err := json.Unmarshal(responseBody, &errResponse); err != nil {
		// 	fmt.Println("Error parsing JSON:", err)
		// 	return
		// }

		// // c.JSON(http.StatusOK, schemeDetails)
		// response := map[string]interface{}{
		// 	"message": errResponse.Message,
		// }
		// c.JSON(http.StatusOK, response)

		return
	}

	// Read and print the response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Response Body:", string(responseBody))
	c.JSON(http.StatusOK, gin.H{"message": "SIP placement done successfully "})

}

// New struct to handle scheme selection request
type SchemeDetailsRequest struct {
	SchemeCode string  `json:"schemeCode"`
	ISIN       string  `json:"isin"`
	OrderType  string  `json:"ordertype"`
	SchemeName string  `json:"schemename"`
	Amount     float64 `json:"amount"`
}

// New struct to respond with scheme details
type SchemeDetailsResponse struct {
	SchemeCode       string  `json:"schemeCode"`
	ISIN             string  `json:"isin"`
	SchemeName       string  `json:"schemeName"`
	CategoryName     string  `json:"categoryName"`
	SubcategoryName  string  `json:"subcategoryName"`
	ReInvestmentPlan string  `json:"reInvestmentPlan"`
	LogoUrl          string  `json:"logoUrl"`
	Returns3yr       float64 `json:"returns3yr"`
	ARQRating        int     `json:"arqRating"`
}

type ErrorResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

// Define the SIP placement request struct
type SIPPlacementRequest struct {
	DpNumber             string  `json:"dpNumber"`
	EmandateId           string  `json:"emandateId"`
	FirstOrderToday      bool    `json:"firstOrderToday"`
	FolioNumber          string  `json:"folioNumber"`
	Frequency            string  `json:"frequency"`
	InstallmentAmount    float64 `json:"installmentAmount"`
	Isin                 string  `json:"isin"`
	NoOfInstallment      int     `json:"noOfInstallment"`
	SchemeCode           string  `json:"schemeCode"`
	StartDate            string  `json:"startDate"`
	TransactionRefNumber string  `json:"transactionRefNumber"`
	Type                 string  `json:"type"`
	VeryFirstSip         bool    `json:"veryFirstSip"`
}
