package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
)

const database = "config"
const coll = "bookings"

var userfields = []string{"email", "startingPoint", "endingPoint", "cabNo"}
var cabFields = []string{"cabNo", "cabLocation"}

func main() {
	log.Println("server is up and running on port 8080")
	http.HandleFunc("/", Index)
	http.HandleFunc("/getpastbookings", GetPastBookings)
	http.HandleFunc("/getcabs", GetCabsNearBy)
	http.HandleFunc("/createcab", CreateCab)
	http.HandleFunc("/createbooking", CreateUserBooking)
	http.ListenAndServe(":8080", nil)
}

//Index ... index page
func Index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Index Page"))
}

//GetPastBookings ... Get List of Bookings of a user
func GetPastBookings(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{}, 0)
	query := make(map[string]interface{}, 0)
	email := r.URL.Query().Get("email")
	if email == "" || len(email) == 0 {
		w.Write([]byte("Please provide mandatory field : email"))
		return
	}
	params["email"] = email
	query["$lt"] = time.Now().AddDate(0, 0, -1).UTC().Format("2006-01-02T15:04:05.000Z")
	params["createdAt"] = query
	result, err := fetchData(params, coll)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	bytesInfo, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytesInfo)
	w.WriteHeader(http.StatusOK)
}

//GetCabsNearBy ...
func GetCabsNearBy(w http.ResponseWriter, r *http.Request) {
	query := make(map[string]interface{}, 0)
	location := r.URL.Query().Get("location")
	if location == "" || len(location) == 0 {
		w.Write([]byte("Please provide mandatory field : location"))
		return
	}
	// query["cabLocation"] = location
	result, err := fetchData(query, "cabs")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if len(result) == 0 {
		w.Write([]byte("No cabs are available"))
		return
	}
	bytesInfo, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytesInfo)
	w.WriteHeader(http.StatusOK)
}

//CreateUserBooking ...
func CreateUserBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("method not alloed"))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userInfo := make(map[string]interface{}, 0)
	decode := json.NewDecoder(r.Body)
	err := decode.Decode(&userInfo)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if checkMandatoryFields(userInfo, userfields) {
		val := fmt.Sprint("Please provide all mandatory fields : ", strings.Join(userfields, ","))
		w.Write([]byte(val))
		return
	}
	if notvalidCab(userInfo) {
		val := fmt.Sprint("Cab details/location mismatch ")
		w.Write([]byte(val))
		return
	}
	defer r.Body.Close()
	err = createData(userInfo, coll)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(" user booking created successfully"))
	w.WriteHeader(http.StatusOK)
}
func checkMandatoryFields(userInfo map[string]interface{}, fields []string) bool {
	for i := 0; i < len(fields); i++ {
		val := userInfo[fields[i]]
		if val == nil || val == "" {
			return true
		}
	}
	return false
}

//dbConnection ...
func dbConnection(db string) (*mgo.Session, error) {
	uri := fmt.Sprintf("mongodb://localhost:27017/%s", db)
	mgoSession, err := mgo.Dial(uri)
	if err == nil {
		return mgoSession, err
	}
	return nil, err
}
func notvalidCab(userInfo map[string]interface{}) bool {
	query := make(map[string]interface{}, 0)
	query["cabLocation"] = userInfo["startingPoint"]
	details, _ := fetchData(query, "cabs")
	if len(details) == 0 {
		return true
	}
	return false
}

//fetchData ... fetch data based on query if query is empty then fetch all the records
func fetchData(params map[string]interface{}, coll string) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)
	session, err := dbConnection(database)
	if err != nil {
		return nil, err
	}
	collection := session.DB(database).C(coll)
	res := collection.Find(params).All(&result)
	if res != nil && res.Error() != "" {
		return nil, errors.New(res.Error())
	}
	defer session.Close()
	return result, nil
}

func createData(userInfo map[string]interface{}, coll string) error {
	session, err := dbConnection(database)
	if err != nil {
		return err
	}
	userInfo["updatedAt"] = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	userInfo["createdAt"] = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	collection := session.DB(database).C(coll)
	err = collection.Insert(userInfo)
	if err != nil {
		return err
	}
	return nil
}

//CreateCab ...
func CreateCab(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("method not alloed"))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cabInfo := make(map[string]interface{}, 0)
	decode := json.NewDecoder(r.Body)
	err := decode.Decode(&cabInfo)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if checkMandatoryFields(cabInfo, cabFields) {
		val := fmt.Sprint("Please provide all mandatory fields : ", strings.Join(cabFields, ","))
		w.Write([]byte(val))
		return
	}
	defer r.Body.Close()
	err = createData(cabInfo, "cabs")
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(" cab created successfully"))
	w.WriteHeader(http.StatusOK)
}
