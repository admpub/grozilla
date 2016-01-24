package main

import (
  "strconv"
  "net/http"
  "log"
  "io/ioutil"
  "sync"
  "os"
  "time"
  "errors"
)

type httpResponse struct {
    resp *http.Response
    err  error
}

const PACKETLENGTH = 32000
var wg sync.WaitGroup
var errorGoRoutine bool

func downloadPacket(client *http.Client, req *http.Request,part_filename string,byteStart, byteEnd int) error {
    c := make(chan httpResponse, 1)
    go func() {
      resp,err := client.Do(req)
      http_response := httpResponse{resp,err}
      c <- http_response
    }()
    select {
    case http_response := <-c:
      if err := handleResponse(http_response,part_filename,byteStart,byteEnd); err != nil {
        return err
      }
    case <-time.After(time.Second * time.Duration(10)):
      err := errors.New("Manual time out as response not recieved")
      return err
    }
    return nil
}

func handleResponse(http_response httpResponse,part_filename string, byteStart, byteEnd int) error {
    if http_response.err != nil{
      return http_response.err
    }
    defer http_response.resp.Body.Close()
    reader, err := ioutil.ReadAll(http_response.resp.Body)
    if err != nil {
      return err
    }
    log.Println(part_filename, len(reader))
    err = writeBytes(part_filename,reader,byteStart,byteEnd)
    if err != nil {
      return err
    }
    return nil
}

func downloadPacketWithRetry(client *http.Client, req *http.Request,part_filename string,byteStart, byteEnd int) error{
  var err error
  for i := 0 ; i < 5 ; i++ {
    err = downloadPacket(client, req, part_filename,byteStart,byteEnd)
    if (err == nil){
      return nil
    } else if (err.Error() == "Manual time out as response not recieved") {
      continue
    } else {
      return err
    }
  }
  return err
}

func downloadPart(url,filename string, index, byteStart, byteEnd int){
    client := &http.Client{}
    part_filename := filename + "_" + strconv.Itoa(index)
    noofpacket := (byteEnd - byteStart + 1)/PACKETLENGTH + 1
    for i := 0 ; i < noofpacket; i ++ {
      packetStart := byteStart + i*PACKETLENGTH
      packetEnd   := packetStart + PACKETLENGTH
      if (i == noofpacket - 1){
        packetEnd = byteEnd
      }
      range_header := "bytes=" + strconv.Itoa(packetStart) +"-" + strconv.Itoa(packetEnd-1)
      //log.Println(range_header)
      req, _ := http.NewRequest("GET",url, nil)
      req.Header.Add("Range", range_header)
      err := downloadPacketWithRetry(client,req,part_filename,byteStart,byteEnd)
      if err != nil {
        handleErrorInGoRoutine(i,err)
        return
      }
      // if (index == 3){
      //     err := errors.New("error in go routine 3")
      //     handleErrorInGoRoutine(i,err)
      //     return
      // }

    }
    wg.Done()
}

func Download(url string,length int){
    partLength := length / *noOfFiles
    filename := getFilenameFromUrl(url)
    for i := 0 ; i < *noOfFiles ; i++ {
      byteStart := partLength * (i)
      byteEnd   := byteStart + partLength
      if (i == *noOfFiles - 1 ){
        byteEnd = length
      }
      os.MkdirAll("temp/", 0777)
      createTempFile("temp/" + filename + "_" + strconv.Itoa(i),byteStart,byteEnd)
      wg.Add(1)
      go downloadPart(url,filename,i,byteStart,byteEnd)
    }
    wg.Wait()
    if (!errorGoRoutine){
      mergeFiles(filename,*noOfFiles)
      clearFiles(filename,*noOfFiles)
      reader,_ := ioutil.ReadFile(filename)
      log.Println(len(reader))
      log.Println("download complete")
    }
}

func Resume(url string,length int){
    filename := getFilenameFromUrl(url)
    *noOfFiles = noOfExistingConnection(filename,length)
    partLength := length / *noOfFiles
    for i := 0 ; i < *noOfFiles ; i++ {
      part_filename := "temp/" +filename + "_" + strconv.Itoa(i)
      if _, err := os.Stat(part_filename); err != nil {
        byteStart := partLength * (i)
        byteEnd   := byteStart + partLength
        if (i == *noOfFiles - 1 ){
          byteEnd = length
        }
        wg.Add(1)
        go downloadPart(url,filename,i,byteStart,byteEnd)
      } else {
        byteStart, byteEnd := readHeader(part_filename)
        if (byteStart < byteEnd) {
          wg.Add(1)
          go downloadPart(url,filename,i,byteStart,byteEnd)
        }
      }
    }
    wg.Wait()
    if (!errorGoRoutine){
      mergeFiles(filename,*noOfFiles)
      clearFiles(filename,*noOfFiles)
      reader,_ := ioutil.ReadFile(filename)
      log.Println(len(reader))
      log.Println("download complete")
    }
}

func DownloadSingle(url string){
    filename := getFilenameFromUrl(url)
    client := &http.Client{}
    req, _ := http.NewRequest("GET",url, nil)
    resp, err := client.Do(req)
    if err != nil {
      log.Fatal(err)
    }
    reader, err := ioutil.ReadAll(resp.Body)
    if err != nil {
      log.Fatal(err)
    }
    log.Println(len(reader))
    err = ioutil.WriteFile(filename, reader,0666)
    if err != nil {
      log.Fatal(err)
    }
}

func handleErrorInGoRoutine(index int, err error){
    log.Println(err)
    errorGoRoutine = true
    wg.Done()
}
