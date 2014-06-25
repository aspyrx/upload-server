package main

import (
    "crypto/md5"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/signal"
    "path/filepath"
    "time"
)

var logFile os.File
var listenAddr string
var outputDir string

func init() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    var logFilePath string
    var userOutputDir string
    flag.StringVar(&logFilePath, "log", "", "Path to log file. Default: no log file")
    flag.StringVar(&listenAddr, "addr", ":80", "Address on which the server listens. Default: :80")
    flag.StringVar(&userOutputDir, "out", "", "Path to output directory for uploaded files. Default: current working directory")

    flag.Parse()

    wd, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }

    // log to file if specified
    if len(logFilePath) > 0 {
        var path string
        if !filepath.IsAbs(logFilePath) {
            path = filepath.Join(wd, logFilePath)
        } else {
            path = logFilePath
        }

        logFile, err := os.OpenFile(path, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0666)
        if err != nil {
            log.Fatal(err)
        }

        log.SetOutput(logFile)
    }

    // set output directory
    if len(userOutputDir) > 0 {
        if !filepath.IsAbs(userOutputDir) {
            outputDir = filepath.Join(wd, userOutputDir)
        } else {
            outputDir = userOutputDir
        }
    } else {
        outputDir = wd
    }
}

func main() {
    defer logFile.Close()

    go httpServer()

    // make a channel to listen for interrupt/kill signals
    c := make(chan os.Signal, 1)
    defer close(c)
    signal.Notify(c, os.Interrupt, os.Kill)

    // block until signal received
    <-c
    log.Println("Stopping...")
}

func httpServer() {
    http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
        // parse the multipart form file upload; store up to 10 MB in memory
        r.ParseMultipartForm(10485760)

        // check that files exist
        if len(r.MultipartForm.File) < 1 {
            return;
        }

        // loop through all FileHeaders
        for _, fhArray := range r.MultipartForm.File {
            for _, fh := range fhArray {
                // open file
                file, err := fh.Open()
                if err != nil {
                    log.Println(err)
                    break
                }
                defer file.Close()

                // compute md5 checksum of file
                h := md5.New()
                _, err = io.Copy(h, file)
                if err != nil {
                    log.Println(err)
                    break
                }

                // seek back to the beginning of the file
                if _, err = file.Seek(0, 0); err != nil {
                    log.Println(err)
                    break
                }

                // get the filename and path to write to
                outFileName := fmt.Sprintf("%s %x %s", time.Now().Format(`2006-01-02 15.04.05 -0700`), h.Sum(nil), fh.Filename)
                outFilePath := filepath.Join(outputDir, outFileName)
                outFile, err := os.OpenFile(outFilePath, os.O_CREATE | os.O_EXCL | os.O_WRONLY, 0666)
                if err != nil {
                    log.Println(err)
                    break
                }
                defer outFile.Close()

                // write the file
                _, err = io.Copy(outFile, file)
            }
        }
    })

    log.Printf("Starting http server on %s", listenAddr)
    if err := http.ListenAndServe(listenAddr, nil); err != nil {
        log.Fatal(err)
    }
}
