package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-spew/spew"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"

	// "net/http/pprof"
	"os"
	"time"

	"github.com/pkg/profile"

	"github.com/gorilla/mux"

	"godotenv"
)

// type Account struct {
// 	Taud int //total audience
// 	Tjournals int
// }

// type Journal struct {
// 	Jtopic string
// 	Jarea string
// 	Jwriter string
// 	country string

// }
// type WriterDetails struct {
// 	Wname string
// 	Country string
// 	Profession string
// 	PastWorksInShort string

// }
// type Book struct {
// 	Bname string
// 	Writer string
// 	WriterDetails
// 	Journal
// 	Audience
// 	Account

// }
// type Audience struct{
// 	name string
// 	country string
// }

type hypotenuse struct {
	trianglePlaneAngle float32 // +ve represent Asset's Hypotenuse & -ve represent Currency's hypotenuse
	Hash               string
	length             int
	MerkelRoot         string
	Index              int
	Timestamp          string
	height             //PrevHash string including genesis block and it contains unique describtor of an organisation
}

type base struct {
	currency int
}

type height struct {
	CommHash string //Community Hash which stays unique
	//OtherSide string
	op         int
	MerkelRoot string
}
type triangle struct { //triangle = Block
	hypotenuse //Hash string
	base       //currency
	OtherSide
}

type OtherSide struct {
	trianglePlaneAngle float32 // +ve represent Asset's Hypotenuse & -ve represent Currency's hypotenuse
	Hash               string
	length             int
	MerkelRoot         string
	Index              int
	Timestamp          string
	height
}

var trianglechain []triangle

func calculatehash(triangle triangle) string {
	record := string(triangle.hypotenuse.Index) + triangle.hypotenuse.Timestamp + string(triangle.base.currency) + triangle.hypotenuse.MerkelRoot + string(triangle.hypotenuse.length) + fmt.Sprintf("%f", triangle.hypotenuse.trianglePlaneAngle) + triangle.hypotenuse.height.CommHash + string(triangle.hypotenuse.height.op) + triangle.hypotenuse.height.MerkelRoot

	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	fmt.Println(hashed)
	//triangle.hypotenuse.hash = hashed , we will be doing in caller function
	return hex.EncodeToString(hashed)
}

func generatetriangle( /*oldtriangle triangle */ oldtriangle triangle, b int) (triangle, error) {
	var newtriangle triangle
	t := time.Now()
	newtriangle.hypotenuse.Index = oldtriangle.hypotenuse.Index + 1
	newtriangle.hypotenuse.Timestamp = t.String()
	newtriangle.hypotenuse.length = oldtriangle.hypotenuse.length + 6
	newtriangle.hypotenuse.MerkelRoot = oldtriangle.hypotenuse.MerkelRoot
	newtriangle.OtherSide.Hash = oldtriangle.hypotenuse.Hash
	newtriangle.OtherSide.Index = oldtriangle.hypotenuse.Index
	newtriangle.OtherSide.length = oldtriangle.hypotenuse.length
	newtriangle.OtherSide.MerkelRoot = oldtriangle.hypotenuse.MerkelRoot
	newtriangle.OtherSide.Timestamp = oldtriangle.hypotenuse.Timestamp
	newtriangle.OtherSide.trianglePlaneAngle = oldtriangle.hypotenuse.trianglePlaneAngle
	newtriangle.base.currency = b
	newtriangle.hypotenuse.Hash = calculatehash(newtriangle)
	fmt.Println(newtriangle.hypotenuse.Hash)
	return newtriangle, nil

}

func istriangleValid(newtriangle, oldtriangle triangle) bool {
	if oldtriangle.hypotenuse.Index+1 != newtriangle.hypotenuse.Index {
		return false
	}
	if oldtriangle.hypotenuse.Hash != newtriangle.OtherSide.Hash {
		return false
	}
	if calculatehash(newtriangle) != newtriangle.hypotenuse.Hash {
		return false
	}

	return true

}

func replaceChain(newtriangles []triangle) {
	if len(newtriangles) > len(trianglechain) {
		trianglechain = newtriangles
	}
}

func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGettrianglechain).Methods("GET")
	muxRouter.HandleFunc("/", handleWritetriangle).Methods("POST")
	return muxRouter
}

func handleGettrianglechain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(trianglechain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

type message struct {
	Currency int
}

func handleWritetriangle(w http.ResponseWriter, r *http.Request) {
	var m message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJson(w, r, http.StatusBadRequest, r.Body)
		return
	}

	defer r.Body.Close()

	newtriangle, err := generatetriangle(trianglechain[len(trianglechain)-1], m.Currency)
	if err != nil {
		respondWithJson(w, r, http.StatusInternalServerError, m)
		return
	}
	if istriangleValid(newtriangle, trianglechain[len(trianglechain)-1]) {
		newtrianglechain := append(trianglechain, newtriangle)
		replaceChain(newtrianglechain)
		spew.Dump(trianglechain)
	}
	respondWithJson(w, r, http.StatusCreated, newtriangle)
}

func respondWithJson(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

// var cpuprofile = flag.String("cpuprofile", "", "./cpu.prof")
// var memprofile = flag.String("memprofile", "", "./mem.prof")

func main() {
	// flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create CPU profile: ", err)
	// 	}
	// 	if err := pprof.StartCPUProfile(f); err != nil {
	// 		log.Fatal("could not start CPU profile: ", err)
	// 	}
	// 	defer pprof.StopCPUProfile()
	// }

	// // ... rest of the program ...

	// if *memprofile != "" {
	// 	f, err := os.Create(*memprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create memory profile: ", err)
	// 	}
	// 	runtime.GC() // get up-to-date statistics
	// 	if err := pprof.WriteHeapProfile(f); err != nil {
	// 		log.Fatal("could not write memory profile: ", err)
	// 	}
	// 	f.Close()
	// }

	defer profile.Start().Stop()
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// bookGenesis:=  new(Book)
	// bookGenesis.Bname= "Ramayan"
	// bookGenesis.WriterDetails.Country="India"
	// bookGenesis.WriterDetails.Wname="Valimiki"
	// bookGenesis.WriterDetails.Profession="Sage"
	// bookGenesis.WriterDetails.PastWorksInShort="Always was a devotee of Lord Rama"
	// bookGenesis.Journal.Jtopic=""
	// bookGenesis.Journal.Jarea=""
	// bookGenesis.Journal.Jwriter=""
	// bookGenesis.Journal.country=""
	// bookGenesis.Audience.country=""
	// bookGenesis.Audience.name=""
	// bookGenesis.Account.Taud=0
	// bookGenesis.Account.Tjournals=0  // abhi is arg ko lena hai batch banane ke samay
	// fmt.Println(bookGenesis)
	// //fmt.Println(calculatehash(triangle{ 1, time.Now().String(), *bookGenesis , "", ""  }))
	// newtriangle.hypotenuse.Index = oldtriangle.hypotenuse.Index +1
	// newtriangle.hypotenuse.Timestamp= t.String()
	// newtriangle.hypotenuse.length= oldtriangle.hypotenuse.length + 6
	// newtriangle.hypotenuse.MerkelRoot= oldtriangle.hypotenuse.MerkelRoot
	go func() {
		t := time.Now()
		var genesisT triangle

		genesisT.hypotenuse.Index = 0
		genesisT.hypotenuse.Timestamp = t.String()
		genesisT.hypotenuse.length = 3
		genesisT.hypotenuse.MerkelRoot = "kumar sandeep"
		genesisT.OtherSide.Hash = "Shiva"
		genesisT.OtherSide.Index = 0
		genesisT.OtherSide.length = 0
		genesisT.OtherSide.MerkelRoot = "shiva"
		genesisT.OtherSide.Timestamp = t.String()
		genesisT.OtherSide.trianglePlaneAngle = 90
		genesisT.base.currency = 3

		genesisT.hypotenuse.Hash = calculatehash(genesisT)
		// genesistriangle , err:= generatetriangle(triangle{ 1, t.String(), *bookGenesis , "", ""  }, )
		// if err!= nil{
		// 	log.Fatal(err)
		// }
		trianglechain = append(trianglechain, genesisT)
		fmt.Println(trianglechain[0])

		//

	}()

	log.Fatal(run())

}
