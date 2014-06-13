// [_Command-line flags_](http://en.wikipedia.org/wiki/Command-line_interface#Command-line_option)
// are a common way to specify options for command-line
// programs. For example, in `wc -l` the `-l` is a
// command-line flag.

package main

// Go provides a `flag` package supporting basic
// command-line flag parsing. We'll use this package to
// implement our example command-line program.
import (
    "flag"
    "fmt"
    "reflect"
    "crypto/tls"
    "net"
    _ "io/ioutil"
    "net/http"
    "net/http/httputil"
    "strings"
    "log"
    "os"
    "os/user"
    "runtime"
    "net/url"
    "bufio"
    "bytes"
    "github.com/bgentry/heroku-go"
)

type Command struct {
}

func (s *Command) Run(client *heroku.Client, app string) {
    log.Println("Not implemented.")
}

type Status struct {
    Command
}

func (s *Status) Run(client *heroku.Client, app string) {
        // https://status.heroku.com/api/v3/current-status.json

        log.Println("=== Heroku Status")
        log.Println("Development: No known issues at this time.")
        log.Println("Production:  No known issues at this time.")
}

type Version struct {
    Command
}

func (s *Version) Run(client *heroku.Client) {
        log.Printf("go-heroku-toolbelt/%s (%s) go/%s\n", "0.001", runtime.GOOS, runtime.Version())
}

type Log struct {
    Command
}

type Credential struct {
    Username string
    Password string
}

func (s *Log) Run(client *heroku.Client, app string ) {


        tail := true
        lines := 10
        options:=&heroku.LogSessionCreateOpts{Tail: &tail, Lines: &lines}

        session, err := client.LogSessionCreate(app, options)
        if err!=nil {
            log.Fatalf ("Error %s", err)
            return
        }

        // X-Heroku-Warning
        u, err := url.Parse(session.LogplexURL)

        c:=make(chan string)

        go func(u *url.URL, c chan string) {

            tcpConn, err := net.Dial("tcp", u.Host + ":443")
            cf := &tls.Config{}
            ssl := tls.Client(tcpConn, cf)

            reader := bufio.NewReader(ssl)
            hc := httputil.NewClientConn(ssl, reader)

            req := &http.Request{
                Method:     "GET",
                URL:        u,
                Proto:      "HTTP/1.1",
                ProtoMajor: 1,
                ProtoMinor: 1,
                Header:     make(http.Header),
                Body:       nil,
                Host:       u.Host,
            }

            req.Header.Add("Host", u.Host)
            req.Header.Add("X-Heroku-API-Version", `2`)
            req.Header.Add("User-Agent", fmt.Sprintf(`go-heroku/0.001 (%s go / %s)`, runtime.GOOS, runtime.Version()))
            req.Header.Add("X-Go-Version", runtime.Version())
            req.Header.Add("X-Go-Platform", runtime.GOOS)

            err = hc.Write(req)

            if err!=nil {
                log.Printf ("Error %s", err)
                return
            }

            _, err = hc.Read(req) 

            if err!=nil {
            }


            for {
                line, err := reader.ReadBytes('\r')

                if err!=nil {
                    log.Printf ("Error %s", err)
                    return
                }

                line = bytes.TrimSpace(line)

                c <- string(line)
            }
        }(u, c)


        for {
            log.Println(<- c)
        }
}

var plugins map[string]reflect.Type 

func registerCommand(name string, v interface{}) {
    rv := reflect.ValueOf(v)
    if rv.Kind() == reflect.Ptr {
        rv = rv.Elem()
    }
    t := rv.Type()

    plugins[name]=t
}

func createClient() (*heroku.Client, error) {
        // parse .netrc file for credentials
        usr, _ := user.Current()

        file, err := os.Open(usr.HomeDir + "/.netrc")
        if err!=nil {
            log.Fatal(err)
        }

        credentials := make(map [string]*Credential)
        machine:= ""

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            if strings.Split(scanner.Text(), " ")[0] == "machine" {
                machine = strings.Split(scanner.Text(), " ")[1]
                credentials[machine]=&Credential{}
            }

            if strings.Split(strings.TrimSpace(scanner.Text()), " ")[0] == "login" {
                credentials[machine].Username = strings.Split(strings.TrimSpace(scanner.Text()), " ")[1]
            }

            if strings.Split(strings.TrimSpace(scanner.Text()), " ")[0] == "password" {
                credentials[machine].Password = strings.Split(strings.TrimSpace(scanner.Text()), " ")[1]
            }
        }

        if err := scanner.Err(); err != nil {
            log.Fatal(err)
        }
        
        username:= credentials["api.heroku.com"].Username
        password:= credentials["api.heroku.com"].Password

        if os.Getenv("HEROKU_API_KEY") != "" {
            password = os.Getenv("HEROKU_API_KEY")
        }

        client := heroku.Client{Username: username, Password: password}
        return &client, nil
}

func main() {
    os.Getenv("HEROKU_API_KEY")

    appPtr := flag.String("app", "", "!    No app specified.")
    flag.Parse()

    plugins = make(map[string]reflect.Type)

    registerCommand("version", &Version{})
    registerCommand("status", &Status{})
    registerCommand("logs", &Log{})

    if (plugins[flag.Args()[0]]!=nil) {
        plugin := reflect.New(plugins[flag.Args()[0]]) 
        pP := plugin.MethodByName("Run")

        client, err := createClient(); 

        if err!=nil {
            log.Fatal(err)
        }

        pP.Call([]reflect.Value{reflect.ValueOf(client), reflect.ValueOf(*appPtr)})
    } else {
        log.Printf("!    `%s` is not a heroku command.", flag.Args()[0])
        // log.Printf("!    Perhaps you meant `logs`.")
        log.Printf("!    See `heroku help` for a list of available commands.")
    }
}

