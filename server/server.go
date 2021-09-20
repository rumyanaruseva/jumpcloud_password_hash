package server

import (
    "crypto/sha512"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "path"
    "strconv"
    "sync"
    "time"
)

// Statistics struct
type Stat struct {
    Total int64 `json:"total"`
    Average int64 `json:"average"`
}

var (
    // Password info
    pwdDelay = 5 * time.Second
    pwdHashedMap = make(map[int64]string)
    pwdHashedCount int64 = 0
    pwdTotalTime int64 = 0
    pwdMutexMap sync.Mutex
    pwdServer http.Server

    // Shutdown info
    shutDown bool = false
    shutdownMutex sync.RWMutex
    shutdownDelay = 1 * time.Second
)

/********************************************************************
HandleRequests()
    Runs the password hash server
    Endpoints:
        /hash  - POST requests to hash a password
        /hash/ - GET requests to retrieve a hashed password by id
        /stats - GET requests for total number of passwords and average time
        /shutdown - GET request to shut the sever down
********************************************************************/
func HandleRequests( port int ) {
    http.HandleFunc( "/", home )
    http.HandleFunc( "/hash", handleHashPost )
    http.HandleFunc( "/hash/", handleHashGet )
    http.HandleFunc( "/stats", handleStats )
    http.HandleFunc( "/shutdown", handleShutDown )
    pwdServer = http.Server{Addr: ":" + strconv.Itoa(port)}
    log.Fatal( pwdServer.ListenAndServe(), nil )
}

/********************************************************************
home()
********************************************************************/
func home( w http.ResponseWriter, r *http.Request ) {
    fmt.Println( "Endpoint: home" )
    fmt.Fprintf( w, "JumpCloud Takehome Assignment - Password Hashing Server!" )
}

/********************************************************************
hashPassword()
    Hashes a password. Returns a base64 encoded string of the SHA512
    hash of the provided password.
********************************************************************/
func hashPassword( password string ) string {

    // Hash the password
    hasher := sha512.New()
    passwordBytes := []byte( password )
    hasher.Write( passwordBytes )
    hashedPassword := hasher.Sum(nil)

    // Convert the hashed password to a base64 encoded string
    base64PasswordHashed := base64.URLEncoding.EncodeToString( hashedPassword )

    return base64PasswordHashed
}

/********************************************************************
delayAndAdd()
    Delays for the specified delay time, hash the password and
    add it to the hashed passwords map.
********************************************************************/
func delayAndAdd( id int64, password string, startTime time.Time ) {

    // Delay the hashing
    time.Sleep( pwdDelay )

    // Hash the password
    hashedPassword := hashPassword( password )
    pwdMutexMap.Lock()
    // Store the password in a map by its id and update the count and total time
    pwdHashedCount++
    pwdHashedMap[ id ] = hashedPassword
    pwdTotalTime += time.Since(startTime).Microseconds()
    pwdMutexMap.Unlock()
}

/********************************************************************
handleHashPost()
    Handles POST requests on the /hash endpoint with a form field
    "password" provding the value to hash. Returns an incrementing
    identifier immediately but the password is not hashed for 5 secs.
********************************************************************/
func handleHashPost( w http.ResponseWriter, r *http.Request ) {
    fmt.Println( "Endpoint: /hash POST" )

    // Check shutdown
    if shutDown {
        fmt.Println( "Server has been shut down!" )
        http.Error( w, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable )
        return
    }

    // Check for POST method
    if r.Method != http.MethodPost {
        fmt.Println( "Only POST requests supported!" )
        http.Error( w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed )
        return
    }

    // Lock the shutdown mutex to ensure the server doesn't
    // shut down while processing this request
    shutdownMutex.RLock()
    defer shutdownMutex.RUnlock()

    // Time the request
    startTime := time.Now()

    // Check for the "password" form field
    password := r.FormValue( "password" )
    if password == "" {
        fmt.Println( "Missing password to hash!" )
        http.Error( w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity )
        return
    }

    // Get the incremented count here, but don't actually increment it yet
    // It'll be incremented when the password is hashed, after the delay
    // This is done so the stats endpoint has accurate average time
    pwdMutexMap.Lock()
    id := pwdHashedCount + 1
    pwdMutexMap.Unlock()

    // Start a go routine to do the wait and add the hashed password
    // to the map, this is done so that the id can be returned right
    // away without the delay
    go delayAndAdd( id, password, startTime )

    // Return the hashed password id
    fmt.Fprintf( w, "%d", id )
}

/********************************************************************
handleHashGet()
    Handles GET requests to retrieve a hashed password by its id.
********************************************************************/
func handleHashGet( w http.ResponseWriter, r *http.Request ) {
    fmt.Println( "Endpoint: /hash/ GET" )

    // Check shutdown
    if shutDown {
        fmt.Println( "Server has been shut down!" )
        http.Error( w, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable )
        return
    }

    // Check for GET method
    if r.Method != http.MethodGet {
        fmt.Println( "Only GET requests supported!" )
        http.Error( w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed )
        return
    }

    // Lock the shutdown mutex to ensure the server doesn't
    // shut down while processing this request
    shutdownMutex.RLock()
    defer shutdownMutex.RUnlock()

    // Get the hashed password, if the provided id exists
    id, _ := strconv.ParseInt( path.Base( r.URL.Path ), 0, 64 )
    pwdMutexMap.Lock()
    hashedPassword := pwdHashedMap[ id ]
    pwdMutexMap.Unlock()

    if hashedPassword == "" {
        fmt.Println( "Passsword id not found!" )
        http.Error( w, http.StatusText(http.StatusNotFound), http.StatusNotFound )
        return
    }

    // Return the hashed password
    fmt.Fprintf( w, hashedPassword )
}

/********************************************************************
handleStats()
    Handles GET requests for basic information about password hashes.
    Current stats:
        Total number of passwords hashed (count of POST requests to the /hash endpoint).
        Average time for processing password hashing requests (in microseconds).
********************************************************************/
func handleStats( w http.ResponseWriter, r *http.Request ) {
    fmt.Println( "Endpoint: /stats" )

    // Check shutdown
    if shutDown {
        fmt.Println( "Server has been shut down!" )
        http.Error( w, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable )
        return
    }

    // Check for GET method
    if r.Method != http.MethodGet {
        fmt.Println( "Only GET requests supported!" )
        http.Error( w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed )
        return
    }

    // Lock the shutdown mutex to ensure the server doesn't
    // shut down while processing this request
    shutdownMutex.RLock()
    defer shutdownMutex.RUnlock()

    // Get the current statistics - total number of requests and average processing time
    pwdMutexMap.Lock()
    total := pwdTotalTime
    count := pwdHashedCount
    pwdMutexMap.Unlock()

    // Don't panic if we get a /stats request before we have any passwords hashed
    if count == 0 {
        fmt.Println( "No hashed passwords yet!" )
        http.Error( w, http.StatusText(http.StatusNotFound), http.StatusNotFound )
        return
    }

    average := total / count
    Stats := Stat{ Total: count, Average: average }

    // Serialize and return the stats
    json.NewEncoder(w).Encode(Stats)
}

/********************************************************************
handleShutDown()
    Handles GET “graceful shutdown request”.
********************************************************************/
func handleShutDown( w http.ResponseWriter, r *http.Request ) {
    fmt.Println( "Endpoint: /shutdown" )

    // Check for GET method
    if r.Method != http.MethodGet {
        fmt.Println( "Only GET requests supported!" )
        http.Error( w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed )
        return
    }

    // Ensure there are no requests currently being processed
    // This is done via a RW mutex
    shutdownMutex.Lock()
    defer shutdownMutex.Unlock()

    shutDown = true

    // Send a shutdown message and delay for a bit
    // so the server can send the message before shutting down
    fmt.Fprintf( w, "Server Shutting Down!" )

	go func() {
		time.Sleep( shutdownDelay )
		err := pwdServer.Shutdown( nil )
        if err != nil {
            fmt.Println( "Server unable to shut down!" )
            http.Error( w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError )
        }
	}()
}