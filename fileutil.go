package main

import (
  "strconv"
  "log"
  "io/ioutil"
  "os"
)

func writeBytes(part_filename string, reader []byte) error{
    err := os.MkdirAll("temp/", 0777)
    if err != nil {
      return err
    }
    file, err := os.OpenFile("temp/" + part_filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE,0666)
    if err != nil {
      return err
    }
    defer file.Close()
    if _, err = file.WriteString(string(reader)); err != nil {
      return err
    }
    return nil
}

func mergeFiles(filename string){
    for i := 0; i < noOfFiles ; i++ {
        part_filename := "temp/" + filename + "_" + strconv.Itoa(i)
        file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND,0666)
        if err != nil {
          log.Fatal(err)
        }
        defer file.Close()
        reader,err := ioutil.ReadFile(part_filename)
        if err != nil {
          log.Fatal(err)
        }
        if _, err = file.WriteString(string(reader)); err != nil {
          log.Fatal(err)
        }
    }
}

func clearFiles(filename string){
  os.RemoveAll("temp")
  // for i := 0; i < noOfFiles ; i++ {
  //   part_filename := "temp/" + filename + "_" + strconv.Itoa(i)
  //   os.Remove(part_filename)
  // }
}